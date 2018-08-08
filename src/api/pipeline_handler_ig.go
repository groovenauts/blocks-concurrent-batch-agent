package api

import (
	"context"
	"net/http"
	"time"

	"models"

	"github.com/labstack/echo"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/check_scaling_task
func (h *PipelineHandler) checkScalingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	if !pl.CanScale() {
		log.Warningf(ctx, "Quit because the pipeline can't scale.\n")
		return c.JSON(http.StatusOK, pl)
	}

	if pl.Cancelled {
		log.Infof(ctx, "Quit because the pipeline is cancelled.\n")
		return c.JSON(http.StatusOK, pl)
	}

	switch {
	case models.StatusesHibernationInProgresss.Include(pl.Status) ||
		models.StatusesHibernating.Include(pl.Status):
		log.Infof(ctx, "Quit because the pipeline is %v so now stopping subscribe_task. \n", pl.Status)
		return c.JSON(http.StatusOK, pl)
	}

	return models.WithScaler(ctx, func(scaler *models.Scaler) error {
		operation, err := scaler.Process(ctx, pl)
		if err != nil {
			return err
		}
		if operation != nil {
			return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
				return PostOperationTask(c, "wait_scaling_task", operation)
			})
		}
		return ReturnJsonWith(c, pl, http.StatusAccepted, func() error {
			return PostPipelineTaskWithETA(c, "check_scaling_task", pl, started.Add(30*time.Second))
		})
	})
}
