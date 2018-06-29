package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// JobController implements the Job resource.
type JobController struct {
	*goa.Controller
}

// NewJobController creates a Job controller.
func NewJobController(service *goa.Service) *JobController {
	return &JobController{Controller: service.NewController("JobController")}
}

// Activate runs the activate action.
func (c *JobController) Activate(ctx *app.ActivateJobContext) error {
	// JobController_Activate: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Activate: end_implement
}

// Create runs the create action.
func (c *JobController) Create(ctx *app.CreateJobContext) error {
	// JobController_Create: start_implement

	// Put your logic here

	return nil
	// JobController_Create: end_implement
}

// Delete runs the delete action.
func (c *JobController) Delete(ctx *app.DeleteJobContext) error {
	// JobController_Delete: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Delete: end_implement
}

// Inactivate runs the inactivate action.
func (c *JobController) Inactivate(ctx *app.InactivateJobContext) error {
	// JobController_Inactivate: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Inactivate: end_implement
}

// Output runs the output action.
func (c *JobController) Output(ctx *app.OutputJobContext) error {
	// JobController_Output: start_implement

	// Put your logic here

	res := &app.JobOutput{}
	return ctx.OK(res)
	// JobController_Output: end_implement
}

// PublishingTask runs the publishing_task action.
func (c *JobController) PublishingTask(ctx *app.PublishingTaskJobContext) error {
	// JobController_PublishingTask: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_PublishingTask: end_implement
}

// Show runs the show action.
func (c *JobController) Show(ctx *app.ShowJobContext) error {
	// JobController_Show: start_implement

	// Put your logic here

	res := &app.Job{}
	return ctx.OK(res)
	// JobController_Show: end_implement
}
