package api

import (
	"fmt"
	"net/http"
	"time"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/check_scaling_task
func (h *PipelineHandler) checkScalingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	if pl.Cancelled {
		switch {
		case models.StatusesOpened.Include(pl.Status):
			log.Infof(ctx, "Pipeline is cancelled.\n")
			return ReturnJsonWith(c, pl, http.StatusNoContent, func() error {
				return PostPipelineTask(c, "close_task", pl)
			})
		case models.StatusesAlreadyClosing.Include(pl.Status):
			log.Warningf(ctx, "Pipeline is cancelled but do nothing because it's already closed or being closed.\n")
			return c.JSON(http.StatusOK, pl)
		default:
			return &models.InvalidStateTransition{
				Msg: fmt.Sprintf("Invalid Pipeline#Status %v to subscribe a Pipeline cancelled", pl.Status),
			}
		}
	}

	switch {
	case models.StatusesHibernationInProgresss.Include(pl.Status) ||
		models.StatusesHibernating.Include(pl.Status):
		log.Infof(ctx, "Pipeline is %v so now stopping subscribe_task. \n", pl.Status)
		return c.JSON(http.StatusOK, pl)
	}

	return models.WithScaler(ctx, func(scaler *models.Scaker) error {
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
