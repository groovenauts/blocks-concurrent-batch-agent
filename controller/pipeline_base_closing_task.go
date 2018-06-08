package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// PipelineBaseClosingTaskController implements the PipelineBaseClosingTask resource.
type PipelineBaseClosingTaskController struct {
	*goa.Controller
}

// NewPipelineBaseClosingTaskController creates a PipelineBaseClosingTask controller.
func NewPipelineBaseClosingTaskController(service *goa.Service) *PipelineBaseClosingTaskController {
	return &PipelineBaseClosingTaskController{Controller: service.NewController("PipelineBaseClosingTaskController")}
}

// Refresh runs the refresh action.
func (c *PipelineBaseClosingTaskController) Refresh(ctx *app.RefreshPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Refresh: start_implement

	// Put your logic here

	res := &app.InstanceGroupOperation{}
	return ctx.OK(res)
	// PipelineBaseClosingTaskController_Refresh: end_implement
}

// Start runs the start action.
func (c *PipelineBaseClosingTaskController) Start(ctx *app.StartPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Start: start_implement

	// Put your logic here

	return nil
	// PipelineBaseClosingTaskController_Start: end_implement
}
