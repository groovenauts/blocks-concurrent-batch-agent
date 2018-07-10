package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// InstanceGroupHealthCheckController implements the InstanceGroupHealthCheck resource.
type InstanceGroupHealthCheckController struct {
	*goa.Controller
}

// NewInstanceGroupHealthCheckController creates a InstanceGroupHealthCheck controller.
func NewInstanceGroupHealthCheckController(service *goa.Service) *InstanceGroupHealthCheckController {
	return &InstanceGroupHealthCheckController{Controller: service.NewController("InstanceGroupHealthCheckController")}
}

// Execute runs the execute action.
func (c *InstanceGroupHealthCheckController) Execute(ctx *app.ExecuteInstanceGroupHealthCheckContext) error {
	// InstanceGroupHealthCheckController_Execute: start_implement

	// Put your logic here

	res := &app.InstanceGroupHealthCheck{}
	return ctx.OK(res)
	// InstanceGroupHealthCheckController_Execute: end_implement
}
