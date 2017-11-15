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

// curl -v -X POST http://localhost:8080/pipelines/1/build_task
func (h *PipelineHandler) buildTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	builder, err := models.NewBuilder(ctx)
	if err != nil {
		return err
	}
	err = builder.Process(ctx, pl)
	if err != nil {
		switch err.(type) {
		case *googleapi.Error:
			e2 := err.(*googleapi.Error)
			switch e2.Code {
			case http.StatusConflict: // googleapi: Error 409: 'projects/optical-hangar-158902/global/deployments/pipeline-mjr-59-20170926-163820' already exists and cannot be created., duplicate
				log.Warningf(ctx, "Skip building because of %v", e2.Message)
				return h.ReturnJsonWith(c, pl, http.StatusNoContent, func() error {
					return h.PostPipelineTask(c, "wait_building_task", pl)
				})
			}
		}
		log.Errorf(ctx, "Failed to build a pipeline %v because of %v\n", pl, err)
		return err
	}

	return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return h.PostPipelineTask(c, "wait_building_task", pl)
	})
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_building_task
func (h *PipelineHandler) waitBuildingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	handler := pl.DeployingHandler(ctx)
	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, handler)
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}

	switch pl.Status {
	case models.Deploying:
		return h.ReturnJsonWith(c, pl, http.StatusAccepted, func() error {
			return h.PostPipelineTaskWithETA(c, "wait_building_task", pl, started.Add(30*time.Second))
		})
	case models.Opened:
		if pl.Cancelled {
			return h.ReturnJsonWith(c, pl, http.StatusNoContent, func() error {
				return h.PostPipelineTask(c, "close_task", pl)
			})
		} else {
			return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
				return h.PostPipelineTask(c, "publish_task", pl)
			})
		}
	default:
		return &models.InvalidStateTransition{Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl)}
	}
}
