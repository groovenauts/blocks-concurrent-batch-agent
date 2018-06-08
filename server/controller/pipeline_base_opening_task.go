package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// PipelineBaseOpeningTaskController implements the PipelineBaseOpeningTask resource.
type PipelineBaseOpeningTaskController struct {
	*goa.Controller
}

// NewPipelineBaseOpeningTaskController creates a PipelineBaseOpeningTask controller.
func NewPipelineBaseOpeningTaskController(service *goa.Service) *PipelineBaseOpeningTaskController {
	return &PipelineBaseOpeningTaskController{Controller: service.NewController("PipelineBaseOpeningTaskController")}
}

// Refresh runs the refresh action.
func (c *PipelineBaseOpeningTaskController) Refresh(ctx *app.RefreshPipelineBaseOpeningTaskContext) error {
	// PipelineBaseOpeningTaskController_Refresh: start_implement

	// Put your logic here

	res := &app.InstanceGroupOperation{}
	return ctx.OK(res)
	// PipelineBaseOpeningTaskController_Refresh: end_implement
}

// Start runs the start action.
func (c *PipelineBaseOpeningTaskController) Start(ctx *app.StartPipelineBaseOpeningTaskContext) error {
	// PipelineBaseOpeningTaskController_Start: start_implement

	// Put your logic here

	return nil
	// PipelineBaseOpeningTaskController_Start: end_implement
}
