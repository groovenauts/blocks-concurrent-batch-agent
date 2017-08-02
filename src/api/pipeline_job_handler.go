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

type JobHandler struct{}

func (h *JobHandler) buildActions() map[string](func(c echo.Context) error) {
	return map[string](func(c echo.Context) error){
		"index":  gae_support.With(plBy("pipeline_id", PlToOrg(withAuth(h.index)))),
		"create": gae_support.With(plBy("pipeline_id", PlToOrg(withAuth(h.create)))),
		"show":   gae_support.With(pjBy("id", PjToPl(PlToOrg(withAuth(h.show))))),
		// "publish": gae_support.With(pjBy("id", PjToPl(PlToOrg(withAuth(h.WaitAndPublish))))),
	}
}

// curl -v -X POST http://localhost:8080/pipelines/3/jobs --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *JobHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	pj := &models.Job{}
	if err := c.Bind(pj); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pl := c.Get("pipeline").(*models.Pipeline)
	pj.Pipeline = pl
	err := pj.CreateAndPublishIfPossible(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create Job: %v\n%v\n", pj, err)
		return err
	}
	log.Debugf(ctx, "Created Job: %v\n", pl)

	if pj.Status == models.Ready {
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/jobs/%s/publish", pj.ID), map[string][]string{})
		t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusCreated, pj)
}

// curl -v http://localhost:8080/pipelines/3/jobs
func (h *JobHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	jobs, err := pl.JobAccessor().All(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jobs)
}

// curl -v http://localhost:8080/jobs/1
func (h *JobHandler) show(c echo.Context) error {
	pj := c.Get("job").(*models.Job)
	return c.JSON(http.StatusOK, pj)
}

// curl -v http://localhost:8080/jobs/1/publish
func (h *JobHandler) WaitAndPublish(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	err := pl.WaitUntil(ctx, models.Opened, 10*time.Second, 5*time.Minute)
	if err != nil {
		return err
	}
	pj := c.Get("job").(*models.Job)
	pj.Status = models.Publishing
	log.Debugf(ctx, "PublishAndUpdate#1: %v\n", pj)
	err = pj.Update(ctx)
	if err != nil {
		return err
	}
	err = pj.PublishAndUpdate(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pj)
}
