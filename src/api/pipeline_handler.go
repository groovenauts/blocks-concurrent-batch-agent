package api

import (
	"fmt"
	"net/http"
	"net/url"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type PipelineHandler struct {
	org_id_name      string
	pipeline_id_name string
}

func (h *PipelineHandler) collection(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(orgBy(h.org_id_name, withAuth(action)))
}

func (h *PipelineHandler) member(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(plBy(h.pipeline_id_name, PlToOrg(withAuth(action))))
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
			err := h.PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
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

// curl -v -X PUT http://localhost:8080/pipelines/1/cancel
// curl -v -X PUT http://localhost:8080/pipelines/1/close
func (h *PipelineHandler) cancel(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	pl.Cancel(ctx)
	switch pl.Status {
	case models.Uninitialized, models.Pending, models.Waiting, models.Reserved:
		pl.Status = models.Closed
		pl.Update(ctx)
		return c.JSON(http.StatusOK, pl)
	case models.Building, models.Deploying:
		// Wait until deploying is finished
		return c.JSON(http.StatusAccepted, pl)
	case models.Opened:
		return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
			return h.PostPipelineTask(c, "close_task", pl)
		})
	case models.Closing, models.ClosingError, models.Closed:
		// Do nothing because it's already closed or being closed
		return c.JSON(http.StatusNoContent, pl)
	default:
		return &models.InvalidStateTransition{
			Msg: fmt.Sprintf("Invalid Pipeline#Status %v to cancel", pl.Status),
		}
	}
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
	return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return h.PostPipelineTask(c, "refresh_task", pl)
	})
}

// curl -v -X	POST http://localhost:8080/pipelines/1/publish_task
func (h *PipelineHandler) publishTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	err := pl.PublishJobs(ctx)
	if err != nil {
		return err
	}
	return h.ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		return h.PostPipelineTask(c, "subscribe_task", pl)
	})
}

// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task
func (h *PipelineHandler) refreshTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	refresher := &models.Refresher{}
	err := refresher.Process(ctx, pl, pl.RefreshHandlerWith(ctx, func(pl *models.Pipeline) error {
		return h.PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
	}))
	if err != nil {
		log.Errorf(ctx, "Failed to refresh pipeline %v because of %v\n", pl, err)
		return err
	}
	return c.JSON(http.StatusOK, pl)
}
