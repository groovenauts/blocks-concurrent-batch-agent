package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// InstanceGroupResizingTaskController implements the InstanceGroupResizingTask resource.
type InstanceGroupResizingTaskController struct {
	*goa.Controller
}

// NewInstanceGroupResizingTaskController creates a InstanceGroupResizingTask controller.
func NewInstanceGroupResizingTaskController(service *goa.Service) *InstanceGroupResizingTaskController {
	return &InstanceGroupResizingTaskController{Controller: service.NewController("InstanceGroupResizingTaskController")}
}

// Start runs the start action.
func (c *InstanceGroupResizingTaskController) Start(ctx *app.StartInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Start: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupResizingTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupResizingTaskController) Watch(ctx *app.WatchInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupResizingTaskController_Watch: end_implement
}
