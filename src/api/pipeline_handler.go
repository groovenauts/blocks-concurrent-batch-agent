package api

import (
	"fmt"
	"net/http"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

type PipelineHandler struct {
	Actions map[string](func(c echo.Context) error)
}

func (h *PipelineHandler) buildActions() {
	h.Actions = map[string](func(c echo.Context) error){
		"index":         gae_support.With(h.withOrg(h.withAuth(h.index))),
		"create":        gae_support.With(h.withOrg(h.withAuth(h.create))),
		"subscriptions": gae_support.With(h.withOrg(h.withAuth(h.subscriptions))),
		"show":          gae_support.With(h.Identified(h.PlToOrg(h.withAuth(h.Identified(h.show))))),
		"close":         gae_support.With(h.Identified(h.PlToOrg(h.withAuth(h.Identified(h.close))))),
		"destroy":       gae_support.With(h.Identified(h.PlToOrg(h.withAuth(h.Identified(h.destroy))))),
		"refresh":       gae_support.With(h.refresh), // Don't use withAuth because this is called from cron
		"refresh_task":  gae_support.With(h.Identified(h.pipelineTask("refresh"))),
		// "build_task": gae_support.With(h.Identified(h.PlToOrg(h.withAuth(h.Identified(h.pipelineTask("build")))))),
		// "close_task": gae_support.With(h.Identified(h.PlToOrg(h.withAuth(h.Identified(h.pipelineTask("close")))))),
	}
}

func (h *PipelineHandler) withOrg(f func(c echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		org_id := c.Param("org_id")
		org, err := models.GlobalOrganizationAccessor.Find(ctx, org_id)
		if err == models.ErrNoSuchOrganization {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "No Organization found for " + org_id})
		}
		if err != nil {
			log.Errorf(ctx, "Failed to find Organization id: %v because of %v\n", org_id, err)
			return err
		}
		c.Set("organization", org)
		return f(c)
	}
}

func (h *PipelineHandler) PlToOrg(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		pl := c.Get("pipeline").(*models.Pipeline)
		pl.LoadOrganization(ctx)
		c.Set("organization", pl.Organization)
		return impl(c)
	}
}

func (h *PipelineHandler) Identified(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param("id")

		var accessor *models.PipelineAccessor
		obj := c.Get("organization")

		if obj == nil {
			accessor = models.GlobalPipelineAccessor
		} else {
			org, ok := c.Get("organization").(*models.Organization)
			if ok {
				accessor = org.PipelineAccessor()
			} else {
				msg := fmt.Sprintf("invalid organization: %v", obj)
				log.Errorf(ctx, "Identified %s\n", msg)
				panic(msg)
			}
		}

		pl, err := accessor.Find(ctx, id)
		switch {
		case err == models.ErrNoSuchPipeline:
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "@Identified %v id: %v\n", err, id)
			return err
		}
		c.Set("pipeline", pl)
		return impl(c)
	}
}

// curl -v -X PUT http://localhost:8080/pipelines/1/close
func (h *PipelineHandler) close(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	id := c.Param("id")
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/close_task", id), map[string][]string{})
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, pl)
}

// curl -v -X POST http://localhost:8080/pipelines/1/build_task
// curl -v -X	POST http://localhost:8080/pipelines/1/close_task
// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task
func (h *PipelineHandler) pipelineTask(action string) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		pl := c.Get("pipeline").(*models.Pipeline)
		err := pl.Process(ctx, action)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, pl)
	}
}

// curl -v -X POST http://localhost:8080/orgs/2/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *PipelineHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	pl := &models.Pipeline{}
	if err := c.Bind(pl); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return err
	}
	org := c.Get("organization").(*models.Organization)
	pl.Organization = org
	err := pl.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create pipeline: %v\n%v\n", pl, err)
		return err
	}
	log.Debugf(ctx, "Created pipeline: %v\n", pl)
	if !pl.Dryrun {
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/build_task", pl.ID), map[string][]string{})
		t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusCreated, pl)
}

// curl -v http://localhost:8080/orgs/2/pipelines
func (h *PipelineHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	org := c.Get("organization").(*models.Organization)
	pipelines, err := org.PipelineAccessor().GetAll(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pipelines)
}

// curl -v http://localhost:8080/orgs/2/pipelines/subscriptions
func (h *PipelineHandler) subscriptions(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	org := c.Get("organization").(*models.Organization)
	subscriptions, err := org.PipelineAccessor().GetActiveSubscriptions(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, subscriptions)
}

// curl -v http://localhost:8080/pipelines/1
func (h *PipelineHandler) show(c echo.Context) error {
	pl := c.Get("pipeline").(*models.Pipeline)
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X DELETE http://localhost:8080/pipelines/1
func (h *PipelineHandler) destroy(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	if err := pl.Destroy(ctx); err != nil {
		switch err.(type) {
		case *models.InvalidOperation:
			res := map[string]interface{}{"message": err.Error()}
			c.JSON(http.StatusNotAcceptable, res)
		default:
			return err
		}
	}
	return c.JSON(http.StatusOK, pl)
}

// This is called from cron
// curl -v -X PUT http://localhost:8080/pipelines/refresh
func (h *PipelineHandler) refresh(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	statuses := map[string]models.Status{"deploying": models.Deploying, "closing": models.Closing}
	res := map[string][]string{}
	for name, st := range statuses {
		orgs, err := models.GlobalOrganizationAccessor.All(ctx)
		if err != nil {
			return err
		}
		for _, org := range orgs {
			ids, err := org.PipelineAccessor().GetIDsByStatus(ctx, st)
			if err != nil {
				return err
			}
			for _, id := range ids {
				t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/refresh_task", id), map[string][]string{})
				if _, err := taskqueue.Add(ctx, t, ""); err != nil {
					return err
				}
			}
			res[org.Name+"-"+name] = ids
		}
	}
	return c.JSON(http.StatusOK, res)
}
