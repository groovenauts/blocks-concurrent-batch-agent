package api

import (
	"net/http"
	"time"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/check_hibernation_task
func (h *PipelineHandler) checkHibernationTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	t, err := time.Parse(time.RFC3339, c.Param("since"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	newTask, err := pl.HasNewTaskSince(ctx, t)
	if err != nil {
		log.Errorf(ctx, "Failed to check new tasks because of %v\n", err)
		return err
	}
	if newTask {
		return c.JSON(http.StatusOK, pl)
	} else {
		return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
			return h.PostPipelineTask(c, "hibernate_task", pl)
		})
	}
}
