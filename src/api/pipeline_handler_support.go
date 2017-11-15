package api

import (
	"fmt"
	"net/url"
	"time"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func (h *PipelineHandler) PostPipelineTaskWith(c echo.Context, action string, pl *models.Pipeline, params url.Values, f func(*taskqueue.Task) error) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/%s", pl.ID, action), params)
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	if f != nil {
		err := f(t)
		if err != nil {
			return err
		}
	}
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		log.Errorf(ctx, "Failed to add a task %v to taskqueue because of %v\n", t, err)
		return err
	}
	return nil
}

func (h *PipelineHandler) SetETAFunc(eta time.Time) func(t *taskqueue.Task) error {
	return func(t *taskqueue.Task) error {
		t.ETA = eta
		return nil
	}
}

func (h *PipelineHandler) PostPipelineTaskWithETA(c echo.Context, action string, pl *models.Pipeline, eta time.Time) error {
	err := h.PostPipelineTaskWith(c, action, pl, url.Values{}, h.SetETAFunc(eta))
	if err != nil {
		return err
	}
	return nil
}

func (h *PipelineHandler) PostPipelineTask(c echo.Context, action string, pl *models.Pipeline) error {
	return h.PostPipelineTaskWithETA(c, action, pl, time.Now())
}

func (h *PipelineHandler) ReturnJsonWith(c echo.Context, pl *models.Pipeline, status int, f func() error) error {
	err := f()
	if err != nil {
		return err
	}
	return c.JSON(status, pl)
}
