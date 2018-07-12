package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// PipelineBaseController implements the PipelineBase resource.
type PipelineBaseController struct {
	*goa.Controller
}

// NewPipelineBaseController creates a PipelineBase controller.
func NewPipelineBaseController(service *goa.Service) *PipelineBaseController {
	return &PipelineBaseController{Controller: service.NewController("PipelineBaseController")}
}

// Close runs the close action.
func (c *PipelineBaseController) Close(ctx *app.ClosePipelineBaseContext) error {
	// PipelineBaseController_Close: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}

		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.PipelineBase) error {
				switch m.Status {
				case model.Hibernating: // Through
				case model.ClosingStarting, model.ClosingRunning, model.ClosingError, model.Closed:
					return ctx.OK(PipelineBaseModelToMediaType(m))
				default:
					return ctx.Conflict(fmt.Errorf("Can't resize because the PipelineBase %q is %s", m.Name, m.Status))
				}

				if err := PostTask(appCtx, fmt.Sprintf("/closing_tasks?resource_id=%d", m.Name), 0); err != nil {
					return err
				}
				return ctx.Created(PipelineBaseModelToMediaType(m))
			})
		}, nil)
	})

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_Close: end_implement
}

// Create runs the create action.
func (c *PipelineBaseController) Create(ctx *app.CreatePipelineBaseContext) error {
	// PipelineBaseController_Create: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		m := PipelineBasePayloadToModel(ctx.Payload)
		m.ParentKey = orgKey
		m.Status = model.OpeningRunning
		err := datastore.RunInTransaction(appCtx, func(c context.Context) error {
			store := &model.PipelineBaseStore{}
			_, err := store.Put(c, &m)
			if err != nil {
				return ctx.BadRequest(goa.ErrBadRequest(err))
			}

			if err := PostTask(appCtx, fmt.Sprintf("/opening_tasks?resource_id=%d", m.Name), 0); err != nil {
				return err
			}
			return nil
		}, nil)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		return ctx.Created(PipelineBaseModelToMediaType(&m))
	})

	// PipelineBaseController_Create: end_implement
}

// Delete runs the delete action.
func (c *PipelineBaseController) Delete(ctx *app.DeletePipelineBaseContext) error {
	// PipelineBaseController_Delete: start_implement

	// Put your logic here

	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.PipelineBase) error {
				switch m.Status {
				case model.OpeningError, model.WakingError, model.HibernationGoingError, model.ClosingError, model.Closed: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't resize because the PipelineBase %q is %s", m.Name, m.Status))
				}

				if err := store.Delete(appCtx, m); err != nil {
					return err
				}

				return ctx.OK(PipelineBaseModelToMediaType(m))
			})
		}, nil)
	})

	// PipelineBaseController_Delete: end_implement
}

// HibernationCheckingTask runs the hibernation_checking_task action.
func (c *PipelineBaseController) HibernationCheckingTask(ctx *app.HibernationCheckingTaskPipelineBaseContext) error {
	// PipelineBaseController_HibernationCheckingTask: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NoContent, func(m *model.PipelineBase) error {
				switch m.Status {
				case model.HibernationChecking: // Through
				case model.Awake, model.HibernationGoing, model.HibernationGoingError:
					return ctx.OK(PipelineBaseModelToMediaType(m))
				default:
					return ctx.NoContent(fmt.Errorf("Can't check hibernation because the PipelineBase %q is %s", m.Name, m.Status))
				}

				m.Status = model.HibernationGoing
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PostTask(appCtx, fmt.Sprintf("/destruction_tasks?resource_id=%d", m.IntanceGroupID), 0); err != nil {
					return err
				}

				return ctx.Created(PipelineBaseModelToMediaType(m))
			})
		}, nil)
	})

	// PipelineBaseController_HibernationCheckingTask: end_implement
}

// HibernationDoneTask runs the hibernation_done_task action.
func (c *PipelineBaseController) HibernationDoneTask(ctx *app.HibernationDoneTaskPipelineBaseContext) error {
	// PipelineBaseController_HibernationDoneTask: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NoContent, func(m *model.PipelineBase) error {
				switch m.Status {
				case model.HibernationGoing: // Through
				case model.Awake, model.HibernationGoingError:
					return ctx.OK(PipelineBaseModelToMediaType(m))
				default:
					return ctx.NoContent(fmt.Errorf("Can't check hibernation because the PipelineBase %q is %s", m.Name, m.Status))
				}

				var respond func(*app.PipelineBase) error
				if ctx.Error == nil || *ctx.Error == "" {
					m.Status = model.HibernationGoingError
					respond = ctx.ResetContent
				} else {
					m.Status = model.Hibernating
					respond = ctx.Accepted
				}

				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				return respond(PipelineBaseModelToMediaType(m))
			})
		}, nil)
	})

	// PipelineBaseController_HibernationDoneTask: end_implement
}

// List runs the list action.
func (c *PipelineBaseController) List(ctx *app.ListPipelineBaseContext) error {
	// PipelineBaseController_List: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}
		models, err := store.All(appCtx)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}

		res := app.PipelineBaseCollection{}
		for _, m := range models {
			res = append(res, PipelineBaseModelToMediaType(m))
		}
		return ctx.OK(res)
	})

	// PipelineBaseController_List: end_implement
}

// PullTask runs the pull_task action.
func (c *PipelineBaseController) PullTask(ctx *app.PullTaskPipelineBaseContext) error {
	// PipelineBaseController_PullTask: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_PullTask: end_implement
}

// Show runs the show action.
func (c *PipelineBaseController) Show(ctx *app.ShowPipelineBaseContext) error {
	// PipelineBaseController_Show: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}
		return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.PipelineBase) error {
			return ctx.OK(PipelineBaseModelToMediaType(m))
		})
	})

	// PipelineBaseController_Show: end_implement
}

// WakeupDoneTask runs the wakeup_done_task action.
func (c *PipelineBaseController) WakeupDoneTask(ctx *app.WakeupDoneTaskPipelineBaseContext) error {
	// PipelineBaseController_WakeupDoneTask: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.PipelineBaseStore{ParentKey: orgKey}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NoContent, func(m *model.PipelineBase) error {
				switch m.Status {
				case model.Waking: // Through
				case model.WakingError, model.Awake:
					return ctx.OK(PipelineBaseModelToMediaType(m))
				default:
					return ctx.NoContent(fmt.Errorf("Can't Wakeup Done because the PipelineBase %q is %s", m.Name, m.Status))
				}

				var respond func(*app.PipelineBase) error
				if ctx.Error == nil || *ctx.Error == "" {
					m.Status = model.WakingError
					respond = ctx.ResetContent
				} else {
					m.Status = model.Awake
					respond = ctx.Accepted
				}

				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				return respond(PipelineBaseModelToMediaType(m))
			})
		}, nil)
	})

	// PipelineBaseController_WakeupDoneTask: end_implement
}
