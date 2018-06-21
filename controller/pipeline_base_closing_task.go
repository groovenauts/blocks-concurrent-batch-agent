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

// Watch runs the watch action.
func (c *PipelineBaseClosingTaskController) Watch(ctx *app.WatchPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Watch: start_implement

	// Put your logic here

	res := &app.Operation{}
	return ctx.OK(res)
	// PipelineBaseClosingTaskController_Watch: end_implement
}
