package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// curl -v -X	POST http://localhost:8080/pipelines/1/subscribe_task
func (h *PipelineHandler) subscribeTask(c echo.Context) error {
	started := time.Now()
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	if pl.Cancelled {
		switch pl.Status {
		case models.Opened:
			log.Infof(ctx, "Pipeline is cancelled.\n")
			return ReturnJsonWith(c, pl, http.StatusNoContent, func() error {
				return PostPipelineTask(c, "close_task", pl)
			})
		case models.Closing, models.ClosingError, models.Closed:
			log.Warningf(ctx, "Pipeline is cancelled but do nothing because it's already closed or being closed.\n")
			return c.JSON(http.StatusOK, pl)
		default:
			return &models.InvalidStateTransition{
				Msg: fmt.Sprintf("Invalid Pipeline#Status %v to subscribe a Pipeline cancelled", pl.Status),
			}
		}
	}

	switch pl.Status {
	case models.HibernationStarting, models.HibernationProcessing, models.HibernationError, models.Hibernating:
		log.Infof(ctx, "Pipeline is %v so now stopping subscribe_task. \n", pl.Status)
		return c.JSON(http.StatusOK, pl)
	}

	err := pl.PullAndUpdateJobStatus(ctx)
	if err != nil {
		switch err.(type) {
		case *models.SubscriprionNotFound:
			switch pl.Status {
			case models.Closing, models.Closed:
				log.Infof(ctx, "Pipeline is already closed\n")
				return c.JSON(http.StatusOK, pl)
			default:
				log.Infof(ctx, "Subscription is not found but the pipeline isn't closed because of %v\n", err)
			}
		default:
			log.Errorf(ctx, "Failed to get Pipeline#PullAndUpdateJobStatus() because of %v\n", err)
			return err
		}
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
			if pl.HibernationDelay == 0 {
				return ReturnJsonWith(c, pl, http.StatusCreated, func() error {
					return PostPipelineTask(c, "close_task", pl)
				})
			} else {
				err := pl.WaitHibernation(ctx)
				if err != nil {
					return err
				}
				now := time.Now()
				eta := now.Add(time.Duration(pl.HibernationDelay) * time.Second)
				params := url.Values{
					"since": []string{now.Format(time.RFC3339)},
				}
				return PostPipelineTaskWith(c, "check_hibernation_task", pl, params, SetETAFunc(eta))
			}
		} else {
			return c.JSON(http.StatusOK, pl)
		}
	} else {
		return ReturnJsonWith(c, pl, http.StatusAccepted, func() error {
			return PostPipelineTaskWithETA(c, "subscribe_task", pl, started.Add(30*time.Second))
		})
	}
}
