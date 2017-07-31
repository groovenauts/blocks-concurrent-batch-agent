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

type PipelineHandler struct {
	Actions map[string](func(c echo.Context) error)
}

func (h *PipelineHandler) buildActions() {
	h.Actions = map[string](func(c echo.Context) error){
		"index":         gae_support.With(orgBy("org_id", withAuth(h.index))),
		"create":        gae_support.With(orgBy("org_id", withAuth(h.create))),
		"subscriptions": gae_support.With(orgBy("org_id", withAuth(h.subscriptions))),
		"show":          gae_support.With(plBy("id", PlToOrg(withAuth(h.show)))),
		"close":         gae_support.With(plBy("id", PlToOrg(withAuth(h.close)))),
		"destroy":       gae_support.With(plBy("id", PlToOrg(withAuth(h.destroy)))),
		"refresh":       gae_support.With(h.refresh), // Don't use withAuth because this is called from cron
		// "refresh_task":  gae_support.With(plBy("id", h.refreshTask)),
		// "build_task": gae_support.With(plBy("id", PlToOrg(withAuth(h.pipelineTask("build"))))),
		// "close_task": gae_support.With(plBy("id", PlToOrg(withAuth(h.pipelineTask("close"))))),
	}
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
	err := pl.ReserveOrWait(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to reserve or wait pipeline: %v\n%v\n", pl, err)
		return err
	}
	log.Debugf(ctx, "Created pipeline: %v\n", pl)
	if pl.Status == models.Reserved {
		if pl.Dryrun {
			log.Debugf(ctx, "[DRYRUN] POST buildTask for %v\n", pl)
		} else {
			t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/build_task", pl.ID), map[string][]string{})
			// build_task checks AUTH_HEADER by using withPlIDHexAuth filter instead of withAuth filter
			// so Set IDHex to AUTH_HEADER instead of original AUTH_HEADER
			t.Header.Add(AUTH_HEADER, fmt.Sprintf("Bearer %s", pl.IDHex()))
			if _, err := taskqueue.Add(ctx, t, ""); err != nil {
				return err
			}
		}
	}
	return c.JSON(http.StatusCreated, pl)
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

// curl -v -X PUT http://localhost:8080/pipelines/1/close
func (h *PipelineHandler) close(c echo.Context) error {
	pl := c.Get("pipeline").(*models.Pipeline)
	return h.PostPipelineTask(c, "close_task", pl, http.StatusOK)
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



// curl -v -X POST http://localhost:8080/pipelines/1/build_task
func (h *PipelineHandler) buildTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	builder, err := models.NewBuilder(ctx)
	if err != nil {
		return err
	}
	err = builder.Process(ctx, pl)
	if err != nil {
		return err
	}

	return h.PostPipelineTask(c, "wait_building_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_building_task
func (h *PipelineHandler) waitBuildingTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	handler := pl.RefreshHandler(ctx)

	for pl.Status == models.Deploying {
		refresher := &models.Refresher{}
		err := refresher.Process(ctx, pl, handler)
		if err != nil {
			return err
		}
		time.Sleep(30 * time.Second)
	}

	return h.PostPipelineTask(c, "publish_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/publish_task
func (h *PipelineHandler) publishTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	accessor := pl.JobAccessor()
	jobs, err := accessor.All(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if job.Status == models.Waiting {
			_, err := job.Publish(ctx)
			if err != nil {
				return err
			}
		}
	}

	return h.PostPipelineTask(c, "subscribe_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/subscribe_task
func (h *PipelineHandler) subscribeTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	for {
		finished, err := pl.AllJobFinished(ctx)
		if err != nil {
			return err
		}
		log.Debugf(ctx, "Pipeline#AllJobFinished() returned %v\n", finished)
		if finished {
			break
		}
		err = pl.PullAndUpdateJobStatus(ctx)
		if err != nil {
			return err
		}
		time.Sleep(30 * time.Second)
	}

	return h.PostPipelineTask(c, "start_closing_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/start_closing_task
func (h *PipelineHandler) startClosingTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx)
	if err != nil {
		return err
	}
	err = closer.Process(ctx, pl)
	if err != nil {
		return err
	}

	return h.PostPipelineTask(c, "wait_closing_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_closing_task
func (h *PipelineHandler) waitClosingTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	pl := c.Get("pipeline").(*models.Pipeline)
	handler := pl.RefreshHandlerWith(ctx, func(pl *models.Pipeline) error {
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/build_task", pl.ID), map[string][]string{})
		t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
		return nil
	})

	for pl.Status == models.Closing {
		refresher := &models.Refresher{}
		err := refresher.Process(ctx, pl, handler)
		if err != nil {
			return err
		}
		time.Sleep(30 * time.Second)
	}

	return c.JSON(http.StatusOK, pl)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/close_task
func (h *PipelineHandler) closeTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx)
	if err != nil {
		return err
	}
	err = closer.Process(ctx, pl)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task
func (h *PipelineHandler) refreshTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, pl.RefreshHandlerWith(ctx, func(pl *models.Pipeline) error {
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/build_task", pl.ID), map[string][]string{})
		// build_task checks AUTH_HEADER by using withPlIDHexAuth filter instead of withAuth filter
		// so Set IDHex to AUTH_HEADER instead of original AUTH_HEADER
		// because build_task is called without AUTH_HEADER.
		t.Header.Add(AUTH_HEADER, fmt.Sprintf("Bearer %s", pl.IDHex()))
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
		return nil
	}))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pl)
}

func (h *PipelineHandler) PostPipelineTask(c echo.Context, action string, pl *models.Pipeline, status int) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/%s", pl.ID, action), map[string][]string{})
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(status, pl)
}
