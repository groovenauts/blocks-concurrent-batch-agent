package api

import (
	"fmt"
	"net/http"
	"time"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/check_hibernation_task
func (h *PipelineHandler) checkHibernationTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	t, err := time.Parse(time.RFC3339, c.Param("since"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	newTask, err := pl.HasNewTaskSince(ctx, t)
	if err != nil {
		log.Errorf(ctx, "Failed to check new tasks because of %v\n", err)
		return err
	}
	if newTask {
		return c.JSON(http.StatusOK, pl)
	} else {
		err := pl.StartHibernation(ctx)
		if err != nil {
			log.Errorf(ctx, "Failed to StartHibernation because of %v\n", err)
			return err
		}
		return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
			return PostPipelineTask(c, "hibernate_task", pl)
		})
	}
}

// curl -v -X	POST http://localhost:8080/pipelines/1/hibernate_task
func (h *PipelineHandler) hibernateTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx, pl.ProcessHibernation)
	if err != nil {
		log.Errorf(ctx, "Failed to create new closer because of %v\n", err)
		return err
	}
	err = closer.Process(ctx, pl)
	if err != nil {
		switch err.(type) {
		case *googleapi.Error:
			e2 := err.(*googleapi.Error)
			switch e2.Code {
			case http.StatusNotFound: // googleapi: Error 404: The object 'projects/optical-hangar-158902/global/deployments/pipeline-mjr-89-20170926-223541' is not found., notFound
				log.Warningf(ctx, "Skip closing because of %v", e2.Message)
				return c.JSON(http.StatusOK, pl)
			}
		}
		log.Errorf(ctx, "Failed to hibernate pipeline because of %v\n", err)
		return err
	}

	return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return PostPipelineTask(c, "wait_hibernation_task", pl)
	})
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_hibernation_task
func (h *PipelineHandler) waitHibernationTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	handler := pl.HibernationHandler(ctx)

	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, handler)
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}

	switch pl.Status {
	case models.HibernationStarting:
		return ReturnJsonWith(c, pl, http.StatusAccepted, func() error {
			return PostPipelineTaskWithETA(c, "wait_hibernation_task", pl, started.Add(30*time.Second))
		})
	case models.Hibernating:
		newTask, err := pl.HasNewTaskSince(ctx, pl.HibernationStartedAt)
		if err != nil {
			log.Errorf(ctx, "Failed to check new tasks because of %v\n", err)
			return err
		}
		if newTask {
			err := pl.BackToReady(ctx)
			if err != nil {
				log.Errorf(ctx, "Failed to BackToReady because of %v\n", err)
				return err
			}
			return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
				return PostPipelineTask(c, "build_task", pl)
			})
		} else {
			return c.JSON(http.StatusOK, pl)
		}
	default:
		return &models.InvalidStateTransition{Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl)}
	}
}
