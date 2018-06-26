package controller

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// IntanceGroupController implements the IntanceGroup resource.
type IntanceGroupController struct {
	*goa.Controller
}

// NewIntanceGroupController creates a IntanceGroup controller.
func NewIntanceGroupController(service *goa.Service) *IntanceGroupController {
	return &IntanceGroupController{Controller: service.NewController("IntanceGroupController")}
}

// Create runs the create action.
func (c *IntanceGroupController) Create(ctx *app.CreateIntanceGroupContext) error {
	// IntanceGroupController_Create: start_implement

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

	// IntanceGroupController_Create: end_implement
}

// Delete runs the delete action.
func (c *IntanceGroupController) Delete(ctx *app.DeleteIntanceGroupContext) error {
	// IntanceGroupController_Delete: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// IntanceGroupController_Delete: end_implement
}

// Destruct runs the destruct action.
func (c *IntanceGroupController) Destruct(ctx *app.DestructIntanceGroupContext) error {
	// IntanceGroupController_Destruct: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// IntanceGroupController_Destruct: end_implement
}

// List runs the list action.
func (c *IntanceGroupController) List(ctx *app.ListIntanceGroupContext) error {
	// IntanceGroupController_List: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)
	store := &model.InstanceGroupStore{}
	models, err := store.GetAll(appCtx)
	if err != nil {
		return ctx.BadRequest(goa.ErrBadRequest(err))
	}

	res := app.InstanceGroupCollection{}
	for _, m := range models {
		res = append(res, InstanceGroupModelToMediaType(m))
	}
	return ctx.OK(res)
	// IntanceGroupController_List: end_implement
}

// Resize runs the resize action.
func (c *IntanceGroupController) Resize(ctx *app.ResizeIntanceGroupContext) error {
	// IntanceGroupController_Resize: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// IntanceGroupController_Resize: end_implement
}

// Show runs the show action.
func (c *IntanceGroupController) Show(ctx *app.ShowIntanceGroupContext) error {
	// IntanceGroupController_Show: start_implement

	// Put your logic here

	res := &app.InstanceGroup{}
	return ctx.OK(res)
	// IntanceGroupController_Show: end_implement
}
