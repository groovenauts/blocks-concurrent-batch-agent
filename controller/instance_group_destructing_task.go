package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// InstanceGroupDestructingTaskController implements the InstanceGroupDestructingTask resource.
type InstanceGroupDestructingTaskController struct {
	*goa.Controller
}

// NewInstanceGroupDestructingTaskController creates a InstanceGroupDestructingTask controller.
func NewInstanceGroupDestructingTaskController(service *goa.Service) *InstanceGroupDestructingTaskController {
	return &InstanceGroupDestructingTaskController{Controller: service.NewController("InstanceGroupDestructingTaskController")}
}

// Watch runs the watch action.
func (c *InstanceGroupDestructingTaskController) Watch(ctx *app.WatchInstanceGroupDestructingTaskContext) error {
	// InstanceGroupDestructingTaskController_Watch: start_implement

	// Put your logic here

	res := &app.Operation{}
	return ctx.OK(res)
	// InstanceGroupDestructingTaskController_Watch: end_implement
}
