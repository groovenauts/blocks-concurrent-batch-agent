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

func PostPipelineTaskWith(c echo.Context, action string, pl *models.Pipeline, params url.Values, f func(*taskqueue.Task) error) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	path := fmt.Sprintf("/pipelines/%s/%s", pl.ID, action)
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

func SetETAFunc(eta time.Time) func(t *taskqueue.Task) error {
	return func(t *taskqueue.Task) error {
		t.ETA = eta
		return nil
	}
}

func PostPipelineTaskWithETA(c echo.Context, action string, pl *models.Pipeline, eta time.Time) error {
	err := PostPipelineTaskWith(c, action, pl, url.Values{}, SetETAFunc(eta))
	if err != nil {
		return err
	}
	return nil
}

func PostPipelineTask(c echo.Context, action string, pl *models.Pipeline) error {
	return PostPipelineTaskWithETA(c, action, pl, time.Now())
}

func ReturnJsonWith(c echo.Context, pl *models.Pipeline, status int, f func() error) error {
	err := f()
	if err != nil {
		return err
	}
	return c.JSON(status, pl)
}
