package api

import (
	"context"
	"net/http"

	"models"

	"github.com/labstack/echo"
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
	operation, err := builder.Process(ctx, pl)
	if err != nil {
		switch err.(type) {
		case *googleapi.Error:
			e2 := err.(*googleapi.Error)
			switch e2.Code {
			case http.StatusConflict: // googleapi: Error 409: 'projects/optical-hangar-158902/global/deployments/pipeline-mjr-59-20170926-163820' already exists and cannot be created., duplicate
				log.Warningf(ctx, "Skip building because of %v", e2.Message)
				return c.JSON(http.StatusNoContent, pl)
			}
		}
		log.Errorf(ctx, "Failed to build a pipeline %v because of %v\n", pl, err)
		return err
	}

	return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return PostOperationTask(c, "wait_building_task", operation)
	})
}
