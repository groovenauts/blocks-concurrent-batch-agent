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

// Methods For Pipeline

func PipelineTaskPath(action string, pl *models.Pipeline) string {
	return fmt.Sprintf("/pipelines/%s/%s", pl.ID, action)
}

func PostPipelineTask(c echo.Context, action string, pl *models.Pipeline) error {
	return PostTask(c, PipelineTaskPath(action, pl))
}

func PostPipelineTaskWithETA(c echo.Context, action string, pl *models.Pipeline, eta time.Time) error {
	return PostTaskWithETA(c, PipelineTaskPath(action, pl), eta)
}

func PostPipelineTaskWith(c echo.Context, action string, pl *models.Pipeline, params url.Values, f func(*taskqueue.Task) error) error {
	return PostTaskWith(c, PipelineTaskPath(action, pl), params, f)
}

// Methods For PipelineOperation

func OperationTaskPath(action string, pl *models.PipelineOperation) string {
	return fmt.Sprintf("/operations/%s/%s", pl.ID, action)
}

func PostOperationTask(c echo.Context, action string, pl *models.PipelineOperation) error {
	return PostTask(c, OperationTaskPath(action, pl))
}

func PostOperationTaskWithETA(c echo.Context, action string, pl *models.PipelineOperation, eta time.Time) error {
	return PostTaskWithETA(c, OperationTaskPath(action, pl), eta)
}

func PostOperationTaskWith(c echo.Context, action string, pl *models.PipelineOperation, params url.Values, f func(*taskqueue.Task) error) error {
	return PostTaskWith(c, OperationTaskPath(action, pl), params, f)
}

// Base methods for tasks

func PostTask(c echo.Context, path string) error {
	return PostTaskWithETA(c, path, time.Now())
}

func PostTaskWithETA(c echo.Context, path string, eta time.Time) error {
	err := PostTaskWith(c, path, url.Values{}, SetETAFunc(eta))
	if err != nil {
		return err
	}
	return nil
}

func SetETAFunc(eta time.Time) func(t *taskqueue.Task) error {
	return func(t *taskqueue.Task) error {
		t.ETA = eta
		return nil
	}
}

func PostTaskWith(c echo.Context, path string, params url.Values, f func(*taskqueue.Task) error) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	t := taskqueue.NewPOSTTask(path, params)
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	if f != nil {
		err := f(t)
		if err != nil {
			log.Errorf(ctx, "Failed to callback because of \v\n", err)
			return err
		}
	}
	log.Debugf(ctx, "taskqueue.Add pipeline task %v\n", path)
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		log.Errorf(ctx, "Failed to add a task %v to taskqueue because of %v\n", t, err)
		return err
	}
	return nil
}

func ReturnJsonWith(c echo.Context, model interface{}, status int, f func() error) error {
	err := f()
	if err != nil {
		return err
	}
	return c.JSON(status, model)
}
