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

// Start runs the start action.
func (c *PipelineBaseOpeningTaskController) Start(ctx *app.StartPipelineBaseOpeningTaskContext) error {
	// PipelineBaseOpeningTaskController_Start: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// PipelineBaseOpeningTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *PipelineBaseOpeningTaskController) Watch(ctx *app.WatchPipelineBaseOpeningTaskContext) error {
	// PipelineBaseOpeningTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// PipelineBaseOpeningTaskController_Watch: end_implement
}
