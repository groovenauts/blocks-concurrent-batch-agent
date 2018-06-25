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

// Start runs the start action.
func (c *PipelineBaseClosingTaskController) Start(ctx *app.StartPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Start: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// PipelineBaseClosingTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *PipelineBaseClosingTaskController) Watch(ctx *app.WatchPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// PipelineBaseClosingTaskController_Watch: end_implement
}
