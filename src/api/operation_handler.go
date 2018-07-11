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
	return gae_support.With(plBy(h.pipeline_id_name, http.StatusNoContent, PlToOrg(withAuth(action))))
}

func (h *OperationHandler) member(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(operationBy(h.operation_id_name, http.StatusNoContent, OperationToPl(PlToOrg(withAuth(action)))))
}

// curl -v -X POST http://localhost:8080/operations/3/wait_building_task --data '' -H 'Content-Type: application/json'
func (h *OperationHandler) waitBuildingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)
	log.Debugf(ctx, "waitBuildingTask #0 operation %v\n", operation)

	err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
		updater := &models.DeploymentUpdater{Servicer: servicer}
		return operation.ProcessDeploy(ctx, updater)
	})
	if err != nil {
		log.Errorf(ctx, "waitBuildingTask #1 operation %v\n", operation)
		return err
	}

	log.Debugf(ctx, "waitBuildingTask #2\n")

	if !operation.Done() {
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_building_task", operation, started.Add(30*time.Second))
		})
	}

	log.Debugf(ctx, "waitBuildingTask #3\n")

	pl, err := operation.LoadPipeline(ctx)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "waitBuildingTask #4\n")
	if pl.Status != models.Opened {
		log.Errorf(ctx, "Invalid state transition: Pipeline must be Opened but %v. pipeline: %v\n", pl.Status, pl)
		return &models.InvalidStateTransition{
			Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl),
		}
	}

	log.Debugf(ctx, "waitBuildingTask #5\n")
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
	log.Debugf(ctx, "waitHibernationTask #0 operation: %v\n", operation)

	err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
		updater := &models.DeploymentUpdater{Servicer: servicer}
		return operation.ProcessHibernation(ctx, updater)
	})
	if err != nil {
		log.Errorf(ctx, "waitHibernationTask #1 operation: %v\n", operation)
		return err
	}

	log.Debugf(ctx, "waitHibernationTask #2\n")
	if !operation.Done() {
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_hibernation_task", operation, started.Add(30*time.Second))
		})
	}

	log.Debugf(ctx, "waitHibernationTask #3\n")
	pl, err := operation.LoadPipeline(ctx)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "waitHibernationTask #4\n")
	switch pl.Status {
	case models.Hibernating:
		if pl.Cancelled {
			log.Debugf(ctx, "waitHibernationTask #5\n")
			err := pl.CloseIfHibernating(ctx)
			if err != nil {
				log.Errorf(ctx, "Failed to CloseAfterHibernation because of %v\n", err)
				return err
			}
			log.Debugf(ctx, "waitHibernationTask #6\n")
			return c.JSON(http.StatusOK, pl)
		}

		log.Debugf(ctx, "waitHibernationTask #7\n")
		newTask, err := pl.HasNewTaskSince(ctx, pl.HibernationStartedAt)
		if err != nil {
			log.Errorf(ctx, "Failed to check new tasks because of %v\n", err)
			return err
		}
		log.Debugf(ctx, "waitHibernationTask #8\n")
		if newTask {
			log.Debugf(ctx, "waitHibernationTask #9\n")
			err := pl.BackToBeReserved(ctx)
			if err != nil {
				log.Errorf(ctx, "Failed to BackToReady because of %v\n", err)
				return err
			}
			log.Debugf(ctx, "waitHibernationTask #10\n")
			return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
				return PostPipelineTask(c, "build_task", pl)
			})
		} else {
			log.Debugf(ctx, "waitHibernationTask #11\n")
			return c.JSON(http.StatusOK, pl)
		}
	case models.HibernationError:
		log.Debugf(ctx, "waitHibernationTask #12\n")
		log.Infof(ctx, "Pipeline is already %v so quit wait_hibernation_task\n", pl.Status)
		return c.JSON(http.StatusOK, pl)
	default:
		log.Debugf(ctx, "waitHibernationTask #13\n")
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
	log.Debugf(ctx, "waitClosingTask #0 operation: %v\n", operation)

	err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
		updater := &models.DeploymentUpdater{Servicer: servicer}
		return operation.ProcessClosing(ctx, updater, func(pl *models.Pipeline) error {
			return PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
		})
	})
	if err != nil {
		log.Errorf(ctx, "waitClosingTask #1 operation: %v\n", operation)
		return err
	}

	log.Debugf(ctx, "waitClosingTask #2\n")
	if !operation.Done() {
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_closing_task", operation, started.Add(30*time.Second))
		})
	}

	log.Debugf(ctx, "waitClosingTask #3\n")
	pl, err := operation.LoadPipeline(ctx)
	if err != nil {
		return err
	}

	switch pl.Status {
	case models.Closed:
		log.Debugf(ctx, "waitClosingTask #4\n")
		return c.JSON(http.StatusOK, pl)
	default:
		log.Debugf(ctx, "waitClosingTask #5\n")
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
	log.Debugf(ctx, "waitScalingTask #0 operation: %v\n", operation)

	return models.WithInstanceGroupServicer(ctx, func(servicer models.InstanceGroupServicer) error {
		log.Debugf(ctx, "waitScalingTask #1\n")
		handler_called := false
		handler := func(_ string) error {
			handler_called = true
			return operation.LoadPipelineWith(ctx, func(pl *models.Pipeline) error {
				return PostPipelineTaskWithETA(c, "check_scaling_task", pl, started.Add(30*time.Second))
			})
		}
		successHandler := func(endTime string) error {
			operation.LoadPipelineWith(ctx, func(pl *models.Pipeline) error {
				pl.LogInstanceSize(ctx, endTime, pl.InstanceSize) // No error is returned
				return nil
			})
			return handler(endTime)
		}

		updater := &models.InstanceGroupUpdater{Servicer: servicer}
		err := updater.Update(ctx, operation, successHandler, handler)
		if err != nil {
			log.Errorf(ctx, "Failed to update operation %v because of %v\n", operation, err)
			return err
		}
		if handler_called {
			log.Debugf(ctx, "waitScalingTask #3\n")
			return c.JSON(http.StatusOK, operation)
		}
		log.Debugf(ctx, "waitScalingTask #4\n")
		return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
			return PostOperationTaskWithETA(c, "wait_scaling_task", operation, started.Add(30*time.Second))
		})
	})
}
