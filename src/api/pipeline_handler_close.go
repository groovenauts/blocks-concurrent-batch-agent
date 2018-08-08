package api

import (
	"context"
	"net/http"

	"models"

	"github.com/labstack/echo"
	"google.golang.org/api/googleapi"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/close_task
func (h *PipelineHandler) closeTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create new closer because of %v\n", err)
		return err
	}
	operation, err := closer.Process(ctx, pl)
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

	err = pl.StartClosing(ctx)
	if err != nil {
		return err
	}

	return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return PostOperationTask(c, "wait_closing_task", operation)
	})
}
