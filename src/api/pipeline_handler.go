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
		"refresh":       gae_support.With(plBy("id", PlToOrg(withAuth(h.refresh)))),
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
	err := pl.CreateWithReserveOrWait(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to reserve or wait pipeline: %v\n%v\n", pl, err)
		return err
	}
	log.Debugf(ctx, "Created pipeline: %v\n", pl)
	err = h.PostPipelineTaskIfPossible(c, pl)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, pl)
}

func (h *PipelineHandler) PostPipelineTaskIfPossible(c echo.Context, pl *models.Pipeline) error {
	if pl.Status == models.Reserved {
		ctx := c.Get("aecontext").(context.Context)
		if pl.Dryrun {
			log.Debugf(ctx, "[DRYRUN] POST buildTask for %v\n", pl)
		} else {
			err := h.PostPipelineTaskWith(c, "build_task", pl, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

// curl -v -X POST http://localhost:8080/pipelines/:id/refresh
func (h *PipelineHandler) refresh(c echo.Context) error {
	pl := c.Get("pipeline").(*models.Pipeline)
	return h.PostPipelineTask(c, "refresh_task", pl, http.StatusOK)
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
		log.Errorf(ctx, "Failed to build a pipeline %v because of %v\n", pl, err)
		return err
	}

	return h.PostPipelineTask(c, "wait_building_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_building_task
func (h *PipelineHandler) waitBuildingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	handler := pl.RefreshHandler(ctx)
	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, handler)
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}

	switch pl.Status {
	case models.Deploying:
		return h.PostPipelineTaskWithETA(c, "wait_building_task", pl, http.StatusNoContent, started.Add(30*time.Second))
	case models.Opened:
		return h.PostPipelineTask(c, "publish_task", pl, http.StatusOK)
	default:
		return &models.InvalidStateTransition{Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl)}
	}
}

// curl -v -X	POST http://localhost:8080/pipelines/1/publish_task
func (h *PipelineHandler) publishTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	err := pl.PublishJobs(ctx)
	if err != nil {
		return err
	}
	return h.PostPipelineTask(c, "subscribe_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/subscribe_task
func (h *PipelineHandler) subscribeTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	err := pl.PullAndUpdateJobStatus(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to get Pipeline#PullAndUpdateJobStatus() because of %v\n", err)
		return err
	}

	jobs, err := pl.JobAccessor().All(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to m.JobAccessor#All() because of %v\n", err)
		return err
	}
	log.Debugf(ctx, "Pipeline has %v jobs\n", len(jobs))

	pendings, err := models.GlobalPipelineAccessor.PendingsFor(ctx, jobs.Finished().IDs())
	if err != nil {
		return err
	}

	for _, pending := range pendings {
		org := c.Get("organization").(*models.Organization)
		pending.Organization = org
		err := pending.UpdateIfReserveOrWait(ctx)
		if err != nil {
			log.Errorf(ctx, "Failed to UpdateIfReserveOrWait pending: %v\n%v\n", pending, err)
			return err
		}
		if pending.Status == models.Reserved {
			err = h.PostPipelineTaskIfPossible(c, pending)
			if err != nil {
				log.Errorf(ctx, "Failed to PostPipelineTaskIfPossible pending: %v\n%v\n", pending, err)
				return err
			}
		}
	}

	if jobs.AllFinished() {
		if pl.ClosePolicy.Match(jobs) {
			return h.PostPipelineTask(c, "start_closing_task", pl, http.StatusOK)
		} else {
			return c.JSON(http.StatusOK, pl)
		}
	} else {
		return h.PostPipelineTaskWithETA(c, "subscribe_task", pl, http.StatusNoContent, started.Add(30*time.Second))
	}
}

// curl -v -X	POST http://localhost:8080/pipelines/1/start_closing_task
func (h *PipelineHandler) startClosingTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create new closer because of %v\n", err)
		return err
	}
	err = closer.Process(ctx, pl)
	if err != nil {
		log.Errorf(ctx, "Failed to close pipeline because of %v\n", err)
		return err
	}

	return h.PostPipelineTask(c, "wait_closing_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/wait_closing_task
func (h *PipelineHandler) waitClosingTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	handler := pl.RefreshHandlerWith(ctx, func(pl *models.Pipeline) error {
		return h.PostPipelineTaskWith(c, "build_task", pl, nil)
	})

	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, handler)
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}

	switch pl.Status {
	case models.Closing:
		return h.PostPipelineTaskWithETA(c, "wait_closing_task", pl, http.StatusNoContent, started.Add(30*time.Second))
	case models.Closed:
		return c.JSON(http.StatusOK, pl)
	default:
		return &models.InvalidStateTransition{Msg: fmt.Sprintf("Unexpected Status: %v for Pipeline: %v", pl.Status, pl)}
	}
}

// curl -v -X	POST http://localhost:8080/pipelines/1/close_task
func (h *PipelineHandler) closeTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	closer, err := models.NewCloser(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create new closer because of %v\n", err)
		return err
	}
	err = closer.Process(ctx, pl)
	if err != nil {
		log.Errorf(ctx, "Failed to close pipeline because of %v\n", err)
		return err
	}

	return h.PostPipelineTask(c, "wait_closing_task", pl, http.StatusOK)
}

// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task
func (h *PipelineHandler) refreshTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, pl.RefreshHandlerWith(ctx, func(pl *models.Pipeline) error {
		return h.PostPipelineTaskWith(c, "build_task", pl, nil)
	}))
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}
	return c.JSON(http.StatusOK, pl)
}

func (h *PipelineHandler) PostPipelineTaskWith(c echo.Context, action string, pl *models.Pipeline, f func(*taskqueue.Task) error) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/%s", pl.ID, action), map[string][]string{})
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

func (h *PipelineHandler) PostPipelineTask(c echo.Context, action string, pl *models.Pipeline, status int) error {
	return h.PostPipelineTaskWithETA(c, action, pl, status, time.Now())
}

func (h *PipelineHandler) PostPipelineTaskWithETA(c echo.Context, action string, pl *models.Pipeline, status int, eta time.Time) error {
	err := h.PostPipelineTaskWith(c, action, pl, func(t *taskqueue.Task) error {
		t.ETA = eta
		return nil
	})
	if err != nil {
		return err
	}
	return c.JSON(status, pl)
}
