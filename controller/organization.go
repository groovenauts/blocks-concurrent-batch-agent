package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// OrganizationController implements the Organization resource.
type OrganizationController struct {
	*goa.Controller
}

// NewOrganizationController creates a Organization controller.
func NewOrganizationController(service *goa.Service) *OrganizationController {
	return &OrganizationController{Controller: service.NewController("OrganizationController")}
}

// Create runs the create action.
func (c *OrganizationController) Create(ctx *app.CreateOrganizationContext) error {
	// OrganizationController_Create: start_implement

	// Put your logic here

	return nil
	// OrganizationController_Create: end_implement
}

// Delete runs the delete action.
func (c *OrganizationController) Delete(ctx *app.DeleteOrganizationContext) error {
	// OrganizationController_Delete: start_implement

	// Put your logic here

	res := &app.Organization{}
	return ctx.OK(res)
	// OrganizationController_Delete: end_implement
}

// List runs the list action.
func (c *OrganizationController) List(ctx *app.ListOrganizationContext) error {
	// OrganizationController_List: start_implement

	// Put your logic here

	res := app.OrganizationCollection{}
	return ctx.OK(res)
	// OrganizationController_List: end_implement
}

// Show runs the show action.
func (c *OrganizationController) Show(ctx *app.ShowOrganizationContext) error {
	// OrganizationController_Show: start_implement

	// Put your logic here

	res := &app.Organization{}
	return ctx.OK(res)
	// OrganizationController_Show: end_implement
}
