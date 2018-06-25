package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// PipelineController implements the Pipeline resource.
type PipelineController struct {
	*goa.Controller
}

// NewPipelineController creates a Pipeline controller.
func NewPipelineController(service *goa.Service) *PipelineController {
	return &PipelineController{Controller: service.NewController("PipelineController")}
}

// Create runs the create action.
func (c *PipelineController) Create(ctx *app.CreatePipelineContext) error {
	// PipelineController_Create: start_implement

	// Put your logic here

	return nil
	// PipelineController_Create: end_implement
}

// Current runs the current action.
func (c *PipelineController) Current(ctx *app.CurrentPipelineContext) error {
	// PipelineController_Current: start_implement

	// Put your logic here

	res := &app.Pipeline{}
	return ctx.OK(res)
	// PipelineController_Current: end_implement
}

// Delete runs the delete action.
func (c *PipelineController) Delete(ctx *app.DeletePipelineContext) error {
	// PipelineController_Delete: start_implement

	// Put your logic here

	res := &app.Pipeline{}
	return ctx.OK(res)
	// PipelineController_Delete: end_implement
}

// List runs the list action.
func (c *PipelineController) List(ctx *app.ListPipelineContext) error {
	// PipelineController_List: start_implement

	// Put your logic here

	res := app.PipelineCollection{}
	return ctx.OK(res)
	// PipelineController_List: end_implement
}

// PreparingFinalizeTask runs the preparing_finalize_task action.
func (c *PipelineController) PreparingFinalizeTask(ctx *app.PreparingFinalizeTaskPipelineContext) error {
	// PipelineController_PreparingFinalizeTask: start_implement

	// Put your logic here

	res := &app.Pipeline{}
	return ctx.OK(res)
	// PipelineController_PreparingFinalizeTask: end_implement
}

// Show runs the show action.
func (c *PipelineController) Show(ctx *app.ShowPipelineContext) error {
	// PipelineController_Show: start_implement

	// Put your logic here

	res := &app.Pipeline{}
	return ctx.OK(res)
	// PipelineController_Show: end_implement
}

// Stop runs the stop action.
func (c *PipelineController) Stop(ctx *app.StopPipelineContext) error {
	// PipelineController_Stop: start_implement

	// Put your logic here

	res := &app.Pipeline{}
	return ctx.OK(res)
	// PipelineController_Stop: end_implement
}
