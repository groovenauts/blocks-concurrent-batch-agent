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

// Refresh runs the refresh action.
func (c *InstanceGroupResizingTaskController) Refresh(ctx *app.RefreshInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Refresh: start_implement

	// Put your logic here

	res := &app.InstanceGroupOperation{}
	return ctx.OK(res)
	// InstanceGroupResizingTaskController_Refresh: end_implement
}

// Start runs the start action.
func (c *InstanceGroupResizingTaskController) Start(ctx *app.StartInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Start: start_implement

	// Put your logic here

	return nil
	// InstanceGroupResizingTaskController_Start: end_implement
}
