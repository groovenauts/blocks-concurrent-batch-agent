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

// curl -v -X POST http://localhost:8080/operations/3/refresh_task --data '' -H 'Content-Type: application/json'
func (h *OperationHandler) RefreshTask(c echo.Context) error {
	operation := c.Get("operation").(*models.PipelineOperation)
	return c.JSON(http.StatusOK, operation)
}
