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

// Refresh runs the refresh action.
func (c *InstanceGroupConstructingTaskController) Refresh(ctx *app.RefreshInstanceGroupConstructingTaskContext) error {
	// InstanceGroupConstructingTaskController_Refresh: start_implement

	// Put your logic here

	res := &app.InstanceGroupOperation{}
	return ctx.OK(res)
	// InstanceGroupConstructingTaskController_Refresh: end_implement
}

// Start runs the start action.
func (c *InstanceGroupConstructingTaskController) Start(ctx *app.StartInstanceGroupConstructingTaskContext) error {
	// InstanceGroupConstructingTaskController_Start: start_implement

	// Put your logic here

	return nil
	// InstanceGroupConstructingTaskController_Start: end_implement
}
