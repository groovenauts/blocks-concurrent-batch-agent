package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"

	"github.com/goadesign/goa"
	"github.com/mjibson/goon"

	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// JobController implements the Job resource.
type JobController struct {
	*goa.Controller
}

// NewJobController creates a Job controller.
func NewJobController(service *goa.Service) *JobController {
	return &JobController{Controller: service.NewController("JobController")}
}

// Activate runs the activate action.
func (c *JobController) Activate(ctx *app.ActivateJobContext) error {
	// JobController_Activate: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)

		// TODO Check if orgKey is included in the ancestors of the key from :id
		store := &model.JobStore{}
		m, err := store.Get(appCtx, ctx.ID)
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.ID, ctx.NotFound, func(m *model.Job) error {
				switch m.Status {
				case model.Inactive: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't activate because the Job %q is %s", m.Id, m.Status))
				}

				m.Status = model.Publishing
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PostTask(appCtx, "/jobs?/"+m.Id+"/publishing_task", 0); err != nil {
					return err
				}
				return ctx.Created(JobModelToMediaType(m))
			})
		}, nil)
	})

	// JobController_Activate: end_implement
}

// Create runs the create action.
func (c *JobController) Create(ctx *app.CreateJobContext) error {
	// JobController_Create: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		if ctx.PipelineBaseID == nil {
			return ctx.BadRequest(goa.ErrBadRequest(fmt.Sprintf("Now pipeline_base_id is required")))
		}
		pbStore := &model.PipelineBaseStore{ParentKey: orgKey}
		pb, err := pbStore.Get(appCtx, *ctx.PipelineBaseID)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		g := goon.FromContext(appCtx)
		key, err := g.KeyError(pb)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}

		m := JobPayloadToModel(ctx.Payload)
		m.Parent = key
		m.Status = model.Inactive
		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			store := &model.JobStore{}
			_, err := store.Put(c, &m)
			if err != nil {
				return ctx.BadRequest(goa.ErrBadRequest(err))
			}
			return ctx.Created(JobModelToMediaType(&m))
		}, nil)
	})

	return nil
	// JobController_Create: end_implement
}

// Delete runs the delete action.
func (c *JobController) Delete(ctx *app.DeleteJobContext) error {
	// JobController_Delete: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Delete: end_implement
}

// Inactivate runs the inactivate action.
func (c *JobController) Inactivate(ctx *app.InactivateJobContext) error {
	// JobController_Inactivate: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Inactivate: end_implement
}

// Output runs the output action.
func (c *JobController) Output(ctx *app.OutputJobContext) error {
	// JobController_Output: start_implement

	// Put your logic here

	res := &app.JobOutput{}
	return ctx.OK(res)
	// JobController_Output: end_implement
}

// PublishingTask runs the publishing_task action.
func (c *JobController) PublishingTask(ctx *app.PublishingTaskJobContext) error {
	// JobController_PublishingTask: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)

	store := &model.JobStore{}
	return c.member(appCtx, store, ctx.ID, ctx.NotFound, func(m *model.Job) error {
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
				log.Errorf(ctx, "Failed to save successfully published message as a Job message: %v, topic: %q, job: %v\n", msg, topic, m)
				return err
			}
		})
	})

	// JobController_PublishingTask: end_implement
}

// Show runs the show action.
func (c *JobController) Show(ctx *app.ShowJobContext) error {
	// JobController_Show: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Show: end_implement
}
