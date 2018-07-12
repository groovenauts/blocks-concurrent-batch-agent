package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// PipelineJobController implements the PipelineJob resource.
type PipelineJobController struct {
	*goa.Controller
}

// NewPipelineJobController creates a PipelineJob controller.
func NewPipelineJobController(service *goa.Service) *PipelineJobController {
	return &PipelineJobController{Controller: service.NewController("PipelineJobController")}
}

// Activate runs the activate action.
func (c *PipelineJobController) Activate(ctx *app.ActivatePipelineJobContext) error {
	// PipelineJobController_Activate: start_implement

	// Put your logic here

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)

		// TODO Check if orgKey is included in the ancestors of the key from :id
		store := &model.JobStore{}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.ID, ctx.BadRequest, ctx.NotFound, func(m *model.Job) error {
				switch m.Status {
				case model.Inactive: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't activate because the Job %q is %s", m.Id, m.Status))
				}

				m.Status = model.Publishing
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PostTask(appCtx, fmt.Sprintf("/jobs/%d/publishing_task", m.Id), 0); err != nil {
					return err
				}
				return ctx.Created(JobModelToMediaType(m))
			})
		}, nil)
	})

	// PipelineJobController_Activate: end_implement
}

// Create runs the create action.
func (c *PipelineJobController) Create(ctx *app.CreatePipelineJobContext) error {
	// PipelineJobController_Create: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		if ctx.PipelineBaseID == nil {
			return ctx.BadRequest(goa.ErrBadRequest(fmt.Sprintf("Now pipeline_base_id is required")))
		}
		pipelineBaseID, err := strconv.ParseInt(*ctx.PipelineBaseID, 10, 64)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(fmt.Errorf("Invalid Pipeline Base ID %q", *ctx.PipelineBaseID)))
		}
		pbStore := &model.PipelineBaseStore{ParentKey: orgKey}
		pb, err := pbStore.Get(appCtx, pipelineBaseID)
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

	// PipelineJobController_Create: end_implement
}

// Delete runs the delete action.
func (c *PipelineJobController) Delete(ctx *app.DeletePipelineJobContext) error {
	// PipelineJobController_Delete: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// PipelineJobController_Delete: end_implement
}

// Inactivate runs the inactivate action.
func (c *PipelineJobController) Inactivate(ctx *app.InactivatePipelineJobContext) error {
	// PipelineJobController_Inactivate: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// PipelineJobController_Inactivate: end_implement
}

// Output runs the output action.
func (c *PipelineJobController) Output(ctx *app.OutputPipelineJobContext) error {
	// PipelineJobController_Output: start_implement

	// Put your logic here

	res := &app.JobOutput{}
	return ctx.OK(res)
	// PipelineJobController_Output: end_implement
}

// Show runs the show action.
func (c *PipelineJobController) Show(ctx *app.ShowPipelineJobContext) error {
	// PipelineJobController_Show: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// PipelineJobController_Show: end_implement
}
