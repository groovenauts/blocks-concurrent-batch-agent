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

	"google.golang.org/appengine/datastore"
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
	log.Debugf(ctx, "waitBuildingTask operation %v\n", operation)

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
			updater := &models.DeploymentUpdater{Servicer: servicer}
			return operation.ProcessDeploy(ctx, updater)
		})
		if err != nil {
			log.Errorf(ctx, "Failed to ProcessDeploy operation %v\n", operation)
			return err
		}

		if !operation.Done() {
			return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
				return PostOperationTaskWithETA(c, "wait_building_task", operation, started.Add(30*time.Second))
			})
		}

		pl := operation.Pipeline
		log.Debugf(ctx, "waitHibernationTask operation done. Pipeline: %v\n", pl)

		if pl.Status != models.Opened {
			log.Errorf(ctx, "Invalid state transition: Pipeline must be Opened but %v. pipeline: %v\n", pl.Status, pl)
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

		return nil
	}, nil)
	if err != nil {
		log.Errorf(ctx, "Error occurred in TX %v\n", err)
		return err
	}
	return nil
}

// curl -v -X	POST http://localhost:8080/operations/1/wait_hibernation_task
func (h *OperationHandler) waitHibernationTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)
	log.Debugf(ctx, "waitHibernationTask operation: %v\n", operation)

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
			updater := &models.DeploymentUpdater{Servicer: servicer}
			return operation.ProcessHibernation(ctx, updater)
		})
		if err != nil {
			log.Errorf(ctx, "Failed to ProcessHibernation operation: %v\n", operation)
			return err
		}

		done := operation.Done()
		log.Debugf(ctx, "waitHibernationTask operation.Status: %v Done => %v\n", operation.Status, done)

		if !done {
			return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
				return PostOperationTaskWithETA(c, "wait_hibernation_task", operation, started.Add(30*time.Second))
			})
		}

		pl := operation.Pipeline
		log.Debugf(ctx, "waitHibernationTask operation done. Pipeline: %v\n", pl)

		switch pl.Status {
		case models.Hibernating:
			if pl.Cancelled {
				err := pl.CloseIfHibernating(ctx)
				if err != nil {
					log.Errorf(ctx, "Failed to CloseAfterHibernation because of %v\n", err)
					return err
				}
				return c.JSON(http.StatusOK, pl)
			}

			newTask, err := pl.HasNewTaskSince(ctx, pl.HibernationStartedAt)
			if err != nil {
				log.Errorf(ctx, "Failed to check new tasks because of %v\n", err)
				return err
			}
			if newTask {
				err := pl.BackToBeReserved(ctx)
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
		case models.HibernationError:
			log.Infof(ctx, "Pipeline is already %v so quit wait_hibernation_task\n", pl.Status)
			return c.JSON(http.StatusOK, pl)
		default:
			return &models.InvalidStateTransition{
				Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl),
			}
		}

		return nil
	}, nil)
	if err != nil {
		log.Errorf(ctx, "Error occurred in TX %v\n", err)
		return err
	}
	return nil
}

// curl -v -X	POST http://localhost:8080/operations/1/wait_closing_task
func (h *OperationHandler) waitClosingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)
	log.Debugf(ctx, "waitClosingTask operation: %v\n", operation)

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := models.WithDefaultDeploymentServicer(ctx, func(servicer models.DeploymentServicer) error {
			updater := &models.DeploymentUpdater{Servicer: servicer}
			return operation.ProcessClosing(ctx, updater, func(pl *models.Pipeline) error {
				return PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
			})
		})
		if err != nil {
			log.Errorf(ctx, "Failed to ProcessClosing operation: %v\n", operation)
			return err
		}

		if !operation.Done() {
			return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
				return PostOperationTaskWithETA(c, "wait_closing_task", operation, started.Add(30*time.Second))
			})
		}

		pl := operation.Pipeline
		log.Debugf(ctx, "waitHibernationTask operation done. Pipeline: %v\n", pl)

		switch pl.Status {
		case models.Closed:
			return c.JSON(http.StatusOK, pl)
		default:
			return &models.InvalidStateTransition{
				Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl),
			}
		}

		return nil
	}, nil)
	if err != nil {
		log.Errorf(ctx, "Error occurred in TX %v\n", err)
		return err
	}
	return nil
}

// curl -v -X	POST http://localhost:8080/operations/1/wait_scaling_task
func (h *OperationHandler) waitScalingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	operation := c.Get("operation").(*models.PipelineOperation)
	log.Debugf(ctx, "waitScalingTask operation: %v\n", operation)

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		return models.WithInstanceGroupServicer(ctx, func(servicer models.InstanceGroupServicer) error {
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
				return c.JSON(http.StatusOK, operation)
			}
			return ReturnJsonWith(c, operation, http.StatusAccepted, func() error {
				return PostOperationTaskWithETA(c, "wait_scaling_task", operation, started.Add(30*time.Second))
			})
		})

		return nil
	}, nil)
	if err != nil {
		log.Errorf(ctx, "Error occurred in TX %v\n", err)
		return err
	}
	return nil
}
