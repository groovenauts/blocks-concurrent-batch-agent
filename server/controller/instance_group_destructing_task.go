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

// Refresh runs the refresh action.
func (c *InstanceGroupDestructingTaskController) Refresh(ctx *app.RefreshInstanceGroupDestructingTaskContext) error {
	// InstanceGroupDestructingTaskController_Refresh: start_implement

	// Put your logic here

	res := &app.InstanceGroupOperation{}
	return ctx.OK(res)
	// InstanceGroupDestructingTaskController_Refresh: end_implement
}

// Start runs the start action.
func (c *InstanceGroupDestructingTaskController) Start(ctx *app.StartInstanceGroupDestructingTaskContext) error {
	// InstanceGroupDestructingTaskController_Start: start_implement

	// Put your logic here

	return nil
	// InstanceGroupDestructingTaskController_Start: end_implement
}
