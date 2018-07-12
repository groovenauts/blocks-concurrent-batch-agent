package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// PipelineBaseJobController implements the PipelineBaseJob resource.
type PipelineBaseJobController struct {
	*goa.Controller
}

// NewPipelineBaseJobController creates a PipelineBaseJob controller.
func NewPipelineBaseJobController(service *goa.Service) *PipelineBaseJobController {
	return &PipelineBaseJobController{Controller: service.NewController("PipelineBaseJobController")}
}

// PublishingTask runs the publishing_task action.
func (c *PipelineBaseJobController) PublishingTask(ctx *app.PublishingTaskPipelineBaseJobContext) error {
	// PipelineBaseJobController_PublishingTask: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)

	store := &model.JobStore{}
	return c.member(appCtx, store, ctx.ID, ctx.BadRequest, ctx.NotFound, func(m *model.Job) error {
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			switch m.Status {
			case model.Publishing: // Through
			default:
				return ctx.Conflict(fmt.Errorf("Can't publish because the Job %q is %s", m.Id, m.Status))
			}

			err := m.Publish(appCtx)
			if err != nil {
				return err
			}

			if _, err := store.Update(appCtx, m); err != nil {
				log.Errorf(ctx, "Failed to save successfully published job message %v\n", m)
				return err
			}

			return nil
		}, nil)
	})

	// PipelineBaseJobController_PublishingTask: end_implement
}
