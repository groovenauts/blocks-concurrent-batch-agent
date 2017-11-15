package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/close_task
func (h *PipelineHandler) closeTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx, pl.StartClosing)
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
		log.Errorf(ctx, "Failed to close pipeline because of %v\n", err)
		return err
	}

	return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return h.PostPipelineTask(c, "wait_closing_task", pl)
	})
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_closing_task
func (h *PipelineHandler) waitClosingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	handler := pl.ClosingHandler(ctx, func(pl *models.Pipeline) error {
		return h.PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
	})

	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, handler)
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}

	switch pl.Status {
	case models.Closing:
		return h.ReturnJsonWith(c, pl, http.StatusAccepted, func() error {
			return h.PostPipelineTaskWithETA(c, "wait_closing_task", pl, started.Add(30*time.Second))
		})
	case models.Closed:
		return c.JSON(http.StatusOK, pl)
	default:
		return &models.InvalidStateTransition{Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl)}
	}
}
