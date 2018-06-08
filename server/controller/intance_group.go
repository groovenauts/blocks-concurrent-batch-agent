package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
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

	return nil
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

	res := app.InstanceGroupCollection{}
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
