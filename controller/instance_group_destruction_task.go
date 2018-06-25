package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// InstanceGroupDestructionTaskController implements the InstanceGroupDestructionTask resource.
type InstanceGroupDestructionTaskController struct {
	*goa.Controller
}

// NewInstanceGroupDestructionTaskController creates a InstanceGroupDestructionTask controller.
func NewInstanceGroupDestructionTaskController(service *goa.Service) *InstanceGroupDestructionTaskController {
	return &InstanceGroupDestructionTaskController{Controller: service.NewController("InstanceGroupDestructionTaskController")}
}

// Start runs the start action.
func (c *InstanceGroupDestructionTaskController) Start(ctx *app.StartInstanceGroupDestructionTaskContext) error {
	// InstanceGroupDestructionTaskController_Start: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupDestructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupDestructionTaskController) Watch(ctx *app.WatchInstanceGroupDestructionTaskContext) error {
	// InstanceGroupDestructionTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupDestructionTaskController_Watch: end_implement
}
