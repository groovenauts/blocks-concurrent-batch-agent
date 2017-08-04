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
		"show":   gae_support.With(jobBy("id", JobToPl(PlToOrg(withAuth(h.show))))),
		// "publish": gae_support.With(jobBy("id", JobToPl(PlToOrg(withAuth(h.WaitAndPublish))))),
	}
}

// curl -v -X POST http://localhost:8080/pipelines/3/jobs --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *JobHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	job := &models.Job{}
	if err := c.Bind(job); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pl := c.Get("pipeline").(*models.Pipeline)
	job.Pipeline = pl
	err := job.CreateAndPublishIfPossible(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create Job: %v\n%v\n", job, err)
		return err
	}
	log.Debugf(ctx, "Created Job: %v\n", pl)

	err = h.StartToWaitAndPublishIfNeeded(c, job)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, job)
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
	job := c.Get("job").(*models.Job)
	return c.JSON(http.StatusOK, job)
}

func (h *JobHandler) StartToWaitAndPublishIfNeeded(c echo.Context, job *models.Job) error {
	if job.Status != models.Ready {
		return nil
	}
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/jobs/%s/publish", job.ID), map[string][]string{})
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return nil
}

// curl -v http://localhost:8080/jobs/1/publish
func (h *JobHandler) WaitAndPublish(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	err := pl.WaitUntil(ctx, models.Opened, 10*time.Second, 5*time.Minute)
	if err != nil {
		return err
	}
	job := c.Get("job").(*models.Job)
	job.Status = models.Publishing
	log.Debugf(ctx, "PublishAndUpdate#1: %v\n", job)
	err = job.Update(ctx)
	if err != nil {
		return err
	}
	err = job.PublishAndUpdate(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, job)
}
