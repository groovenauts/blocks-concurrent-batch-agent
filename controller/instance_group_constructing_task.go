package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// InstanceGroupConstructingTaskController implements the InstanceGroupConstructingTask resource.
type InstanceGroupConstructingTaskController struct {
	*goa.Controller
}

// NewInstanceGroupConstructingTaskController creates a InstanceGroupConstructingTask controller.
func NewInstanceGroupConstructingTaskController(service *goa.Service) *InstanceGroupConstructingTaskController {
	return &InstanceGroupConstructingTaskController{Controller: service.NewController("InstanceGroupConstructingTaskController")}
}

// Watch runs the watch action.
func (c *InstanceGroupConstructingTaskController) Watch(ctx *app.WatchInstanceGroupConstructingTaskContext) error {
	// InstanceGroupConstructingTaskController_Watch: start_implement

	// Put your logic here

	res := &app.Operation{}
	return ctx.OK(res)
	// InstanceGroupConstructingTaskController_Watch: end_implement
}
