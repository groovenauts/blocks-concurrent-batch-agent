package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// InstanceGroupConstructionTaskController implements the InstanceGroupConstructionTask resource.
type InstanceGroupConstructionTaskController struct {
	*goa.Controller
}

// NewInstanceGroupConstructionTaskController creates a InstanceGroupConstructionTask controller.
func NewInstanceGroupConstructionTaskController(service *goa.Service) *InstanceGroupConstructionTaskController {
	return &InstanceGroupConstructionTaskController{Controller: service.NewController("InstanceGroupConstructionTaskController")}
}

// Start runs the start action.
func (c *InstanceGroupConstructionTaskController) Start(ctx *app.StartInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Start: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupConstructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupConstructionTaskController) Watch(ctx *app.WatchInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupConstructionTaskController_Watch: end_implement
}
