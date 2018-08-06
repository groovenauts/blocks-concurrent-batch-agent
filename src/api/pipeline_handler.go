package api

import (
	"fmt"
	"net/http"
	"net/url"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type PipelineHandler struct {
	org_id_name      string
	pipeline_id_name string
}

func (h *PipelineHandler) collection(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(orgBy(h.org_id_name, http.StatusNotFound, withAuth(action)))
}

func (h *PipelineHandler) member(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(plBy(h.pipeline_id_name, http.StatusNotFound, PlToOrg(withAuth(action))))
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
			err := PostPipelineTaskWith(c, "build_task", pl, url.Values{}, nil)
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
	st := pl.Status
	pl.Cancel(ctx)
	switch {
	case models.StatusesNotDeployedYet.Include(st):
		return c.JSON(http.StatusOK, pl)
	case models.StatusesNowDeploying.Include(st):
		// Wait until deploying is finished
		return c.JSON(http.StatusAccepted, pl)
	case models.StatusesOpened.Include(st):
		return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
			return PostPipelineTask(c, "close_task", pl)
		})
	case models.StatusesAlreadyClosing.Include(st):
		// Do nothing because it's already closed or being closed
		return c.JSON(http.StatusNoContent, pl)
	case models.StatusesHibernationInProgresss.Include(st):
		// Do nothing because it's already started hibernation
		return c.JSON(http.StatusNoContent, pl)
	case models.StatusesHibernating.Include(st):
		return c.JSON(http.StatusOK, pl)
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

// curl -v -X	POST http://localhost:8080/pipelines/1/publish_task
func (h *PipelineHandler) publishTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	err := pl.PublishJobs(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to Pipeline.PublishJobs for %v because of %v\n", pl.ID, err)
		return err
	}
	return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
		if pl.CanScale() {
			err = PostPipelineTask(c, "check_scaling_task", pl)
			if err != nil {
				return err
			}
		}
		jobCount, err := pl.JobCount(ctx, models.Publishing, models.Published, models.Executing)
		if err != nil {
			return err
		}
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			return pl.CalcAndUpdatePullingTaskSize(ctx, jobCount, func(newTasks int) error {
				for i := 0; i < newTasks; i++ {
					if err := PostPipelineTask(c, "subscribe_task", pl); err != nil {
						log.Warningf(ctx, "Failed to start subscribe_task for %v because of %v\n", pl.ID, err)
						// return err
					}
				}
				return nil
			})
		}, nil)
		if err != nil {
			log.Errorf(ctx, "Error occurred in TX %v", err)
			return err
		}
		return nil
	})
}
