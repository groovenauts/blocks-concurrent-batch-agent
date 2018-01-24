package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type OperationHandler struct {
	pipeline_id_name  string
	operation_id_name string
}

func (h *OperationHandler) collection(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(plBy(h.pipeline_id_name, PlToOrg(withAuth(action))))
}

func (h *OperationHandler) member(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(operationBy(h.operation_id_name, OperationToPl(PlToOrg(withAuth(action)))))
}

// curl -v -X POST http://localhost:8080/operations/3/wait_building_task --data '' -H 'Content-Type: application/json'
func (h *OperationHandler) waitBuildingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)

	err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
		updater := &models.DeploymentUpdater{Servicer: servicer}
		return operation.ProcessDeploy(ctx, updater)
	})
	if err != nil {
		return err
	}

	if !operation.Done() {
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_building_task", operation, started.Add(30*time.Second))
		})
	}

	pl, err := operation.LoadPipeline(ctx)
	if err != nil {
		return err
	}
	if pl.Status != models.Opened {
		return &models.InvalidStateTransition{
			Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl),
		}
	}

	if pl.Cancelled {
		return ReturnJsonWith(c, pl, http.StatusNoContent, func() error {
			return PostPipelineTask(c, "close_task", pl)
		})
	} else {
		return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
			return PostPipelineTask(c, "publish_task", pl)
		})
	}
}

// curl -v -X	POST http://localhost:8080/operations/1/wait_hibernation_task
func (h *OperationHandler) waitHibernationTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)

	log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 0)

	err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
		updater := &models.DeploymentUpdater{Servicer: servicer}
		return operation.ProcessHibernation(ctx, updater)
	})
	if err != nil {
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 1)
		return err
	}

	log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 2)

	if !operation.Done() {
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 3)
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_hibernation_task", operation, started.Add(30*time.Second))
		})
	}

	log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 4)

	pl, err := operation.LoadPipeline(ctx)
	if err != nil {
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 5)
		return err
	}

	log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 6)

	switch pl.Status {
	case models.Hibernating:
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 7)
		if pl.Cancelled {
			log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 8)
			err := pl.CloseIfHibernating(ctx)
			if err != nil {
				log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 9)
				log.Errorf(ctx, "Failed to CloseAfterHibernation because of %v\n", err)
				return err
			}
			log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 10)
			return c.JSON(http.StatusOK, pl)
		}
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 11)

		newTask, err := pl.HasNewTaskSince(ctx, pl.HibernationStartedAt)
		if err != nil {
			log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 12)
			log.Errorf(ctx, "Failed to check new tasks because of %v\n", err)
			return err
		}
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 13)
		if newTask {
			log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 14)
			err := pl.BackToBeReserved(ctx)
			if err != nil {
				log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 15)
				log.Errorf(ctx, "Failed to BackToReady because of %v\n", err)
				return err
			}
			log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 16)
			return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
				return PostPipelineTask(c, "build_task", pl)
			})
		} else {
			log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 17)
			return c.JSON(http.StatusOK, pl)
		}
	case models.HibernationError:
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 18)
		log.Infof(ctx, "Pipeline is already %v so quit wait_hibernation_task\n", pl.Status)
		return c.JSON(http.StatusOK, pl)
	default:
		log.Debugf(ctx, "OperationHandler#waitHibernationTask %d\n", 19)
		return &models.InvalidStateTransition{
			Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl),
		}
	}
}

// curl -v -X	POST http://localhost:8080/operations/1/wait_closing_task
func (h *OperationHandler) waitClosingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)

	err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
		updater := &models.DeploymentUpdater{Servicer: servicer}
		return operation.ProcessClosing(ctx, updater, func(pl *models.Pipeline) error {
			return PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
		})
	})
	if err != nil {
		return err
	}

	if !operation.Done() {
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_closing_task", operation, started.Add(30*time.Second))
		})
	}

	pl, err := operation.LoadPipeline(ctx)
	if err != nil {
		return err
	}

	switch pl.Status {
	case models.Closed:
		return c.JSON(http.StatusOK, pl)
	default:
		return &models.InvalidStateTransition{
			Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl),
		}
	}
}

// curl -v -X	POST http://localhost:8080/operations/1/wait_scaling_task
func (h *OperationHandler) waitScalingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)

	err := models.WithInstanceGroupServicer(ctx, func(servicer models.InstanceGroupServicer) error {
		handler := func() error {
			return operation.LoadPipelineWith(ctx, func(pl *models.Pipeline) error {
				return PostPipelineTaskWithETA(c, "check_scaling_task", pl, started.Add(30*time.Second))
			})
		}
		updater := &models.InstanceGroupUpdater{Servicer: servicer}
		return updater.Update(ctx, operation, handler, handler)
	})
	if err != nil {
		return err
	}

	return nil
}
