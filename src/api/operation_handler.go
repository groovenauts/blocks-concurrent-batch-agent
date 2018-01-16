package api

import (
	"fmt"
	"net/http"
	"time"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
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

	err := WithDefaultDeploymentServicer(func(servicer DeploymentServicer) error {
		updater := &DeploymentUpdater{servicer: servicer}
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

	pl, err := operation.LoadPipeline()
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
