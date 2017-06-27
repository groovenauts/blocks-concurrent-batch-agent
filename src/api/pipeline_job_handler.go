package api

import (
	"net/http"
	"time"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type PipelineJobHandler struct {}

func (h *PipelineJobHandler) buildActions() map[string](func(c echo.Context) error) {
	return map[string](func(c echo.Context) error){
		"index":  gae_support.With(plBy("pipeline_id", PlToOrg(withAuth(h.index)))),
		"create": gae_support.With(plBy("pipeline_id", PlToOrg(withAuth(h.create)))),
		"show":		gae_support.With(pjBy("id", PjToPl(PlToOrg(withAuth(h.show))))),
		"publish": gae_support.With(pjBy("id", PjToPl(PlToOrg(withAuth(h.WaitAndPublish))))),
	}
}

// curl -v -X POST http://localhost:8080/orgs/2/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *PipelineJobHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	pj := &models.PipelineJob{}
	if err := c.Bind(pj); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return err
	}
	pl := c.Get("pipeline").(*models.Pipeline)
	pj.Pipeline = pl
	err := pj.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create pipelineJob: %v\n%v\n", pj, err)
		return err
	}
	log.Debugf(ctx, "Created pipelineJob: %v\n", pl)
	return c.JSON(http.StatusCreated, pj)
}

// curl -v http://localhost:8080/orgs/2/pipelines
func (h *PipelineJobHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	jobs, err := pl.JobAccessor().All(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jobs)
}

// curl -v http://localhost:8080/pipelines/1
func (h *PipelineJobHandler) show(c echo.Context) error {
	pj := c.Get("pipeline_job").(*models.PipelineJob)
	return c.JSON(http.StatusOK, pj)
}

// curl -v http://localhost:8080/pipelines/1
func (h *PipelineJobHandler) WaitAndPublish(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	err := pl.WaitUntil(ctx, models.Opened, 10 * time.Second, 5 * time.Minute)
	if err != nil {
		return err
	}
	pj := c.Get("pipeline_job").(*models.PipelineJob)
	err = pj.PublishAndUpdate(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pj)
}
