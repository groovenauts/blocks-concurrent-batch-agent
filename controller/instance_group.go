package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// InstanceGroupController implements the InstanceGroup resource.
type InstanceGroupController struct {
	*goa.Controller
}

// NewInstanceGroupController creates a InstanceGroup controller.
func NewInstanceGroupController(service *goa.Service) *InstanceGroupController {
	return &InstanceGroupController{Controller: service.NewController("InstanceGroupController")}
}

// Create runs the create action.
func (c *InstanceGroupController) Create(ctx *app.CreateInstanceGroupContext) error {
	// InstanceGroupController_Create: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		m := InstanceGroupPayloadToModel(ctx.Payload)
		m.Parent = orgKey
		m.Status = model.ConstructionStarting
		err := datastore.RunInTransaction(appCtx, func(c context.Context) error {
			store := &model.InstanceGroupStore{}
			_, err := store.Put(c, &m)
			if err != nil {
				return ctx.BadRequest(goa.ErrBadRequest(err))
			}

			task := &taskqueue.Task{Path: "/construction_tasks?resource_id=" + m.Id}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return nil
		}, nil)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		return ctx.Created(InstanceGroupModelToMediaType(&m))
	})

	// InstanceGroupController_Create: end_implement
}

// Delete runs the delete action.
func (c *InstanceGroupController) Delete(ctx *app.DeleteInstanceGroupContext) error {
	// InstanceGroupController_Delete: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// InstanceGroupController_Delete: end_implement
}

// Destruct runs the destruct action.
func (c *InstanceGroupController) Destruct(ctx *app.DestructInstanceGroupContext) error {
	// InstanceGroupController_Destruct: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// InstanceGroupController_Destruct: end_implement
}

// List runs the list action.
func (c *InstanceGroupController) List(ctx *app.ListInstanceGroupContext) error {
	// InstanceGroupController_List: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}
		models, err := store.GetAll(appCtx)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}

		res := app.InstanceGroupCollection{}
		for _, m := range models {
			res = append(res, InstanceGroupModelToMediaType(m))
		}
		return ctx.OK(res)
	})

	// InstanceGroupController_List: end_implement
}

// Resize runs the resize action.
func (c *InstanceGroupController) Resize(ctx *app.ResizeInstanceGroupContext) error {
	// InstanceGroupController_Resize: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}

		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			m, err := store.Get(appCtx, ctx.ID)
			if err != nil {
				if err == datastore.ErrNoSuchEntity {
					return ctx.NotFound(fmt.Errorf("InstanceGroup not found: %q", m.Id))
				} else {
					return err
				}
			}

			switch m.Status {
			case model.Constructed, model.ResizeStarting, model.ResizeRunning: // Through
			default:
				return ctx.Conflict(fmt.Errorf("Can't resize because the InstanceGroup %q is %s", m.Id, m.Status))
			}

			if m.InstanceSizeRequested >= ctx.NewSize {
				return ctx.OK(InstanceGroupModelToMediaType(m))
			}

			m.InstanceSizeRequested = ctx.NewSize
			if _, err := store.Update(appCtx, m); err != nil {
				return err
			}

			task := &taskqueue.Task{Path: "/resizing_tasks?resource_id=" + m.Id}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return ctx.Created(InstanceGroupModelToMediaType(m))
		}, nil)
	})

	// InstanceGroupController_Resize: end_implement
}

// Show runs the show action.
func (c *InstanceGroupController) Show(ctx *app.ShowInstanceGroupContext) error {
	// InstanceGroupController_Show: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// InstanceGroupController_Show: end_implement
}
