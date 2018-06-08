package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// PipelineBaseController implements the PipelineBase resource.
type PipelineBaseController struct {
	*goa.Controller
}

// NewPipelineBaseController creates a PipelineBase controller.
func NewPipelineBaseController(service *goa.Service) *PipelineBaseController {
	return &PipelineBaseController{Controller: service.NewController("PipelineBaseController")}
}

// Close runs the close action.
func (c *PipelineBaseController) Close(ctx *app.ClosePipelineBaseContext) error {
	// PipelineBaseController_Close: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_Close: end_implement
}

// Create runs the create action.
func (c *PipelineBaseController) Create(ctx *app.CreatePipelineBaseContext) error {
	// PipelineBaseController_Create: start_implement

	// Put your logic here

	return nil
	// PipelineBaseController_Create: end_implement
}

// Delete runs the delete action.
func (c *PipelineBaseController) Delete(ctx *app.DeletePipelineBaseContext) error {
	// PipelineBaseController_Delete: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_Delete: end_implement
}

// HibernationCheckingFinalizeTask runs the hibernation_checking_finalize_task action.
func (c *PipelineBaseController) HibernationCheckingFinalizeTask(ctx *app.HibernationCheckingFinalizeTaskPipelineBaseContext) error {
	// PipelineBaseController_HibernationCheckingFinalizeTask: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_HibernationCheckingFinalizeTask: end_implement
}

// HibernationGoingFinalizeTask runs the hibernation_going_finalize_task action.
func (c *PipelineBaseController) HibernationGoingFinalizeTask(ctx *app.HibernationGoingFinalizeTaskPipelineBaseContext) error {
	// PipelineBaseController_HibernationGoingFinalizeTask: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_HibernationGoingFinalizeTask: end_implement
}

// List runs the list action.
func (c *PipelineBaseController) List(ctx *app.ListPipelineBaseContext) error {
	// PipelineBaseController_List: start_implement

	// Put your logic here

	res := app.PipelineBaseCollection{}
	return ctx.OK(res)
	// PipelineBaseController_List: end_implement
}

// PullTask runs the pull_task action.
func (c *PipelineBaseController) PullTask(ctx *app.PullTaskPipelineBaseContext) error {
	// PipelineBaseController_PullTask: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_PullTask: end_implement
}

// Show runs the show action.
func (c *PipelineBaseController) Show(ctx *app.ShowPipelineBaseContext) error {
	// PipelineBaseController_Show: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_Show: end_implement
}

// WakingFinalizeTask runs the waking_finalize_task action.
func (c *PipelineBaseController) WakingFinalizeTask(ctx *app.WakingFinalizeTaskPipelineBaseContext) error {
	// PipelineBaseController_WakingFinalizeTask: start_implement

	// Put your logic here

	res := &app.PipelineBase{}
	return ctx.OK(res)
	// PipelineBaseController_WakingFinalizeTask: end_implement
}
