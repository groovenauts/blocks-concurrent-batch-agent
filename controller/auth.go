package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// AuthController implements the Auth resource.
type AuthController struct {
	*goa.Controller
}

// NewAuthController creates a Auth controller.
func NewAuthController(service *goa.Service) *AuthController {
	return &AuthController{Controller: service.NewController("AuthController")}
}

// Create runs the create action.
func (c *AuthController) Create(ctx *app.CreateAuthContext) error {
	// AuthController_Create: start_implement

	// Put your logic here

	return nil
	// AuthController_Create: end_implement
}

// Delete runs the delete action.
func (c *AuthController) Delete(ctx *app.DeleteAuthContext) error {
	// AuthController_Delete: start_implement

	// Put your logic here

	res := &app.Auth{}
	return ctx.OK(res)
	// AuthController_Delete: end_implement
}

// List runs the list action.
func (c *AuthController) List(ctx *app.ListAuthContext) error {
	// AuthController_List: start_implement

	// Put your logic here

	res := app.AuthCollection{}
	return ctx.OK(res)
	// AuthController_List: end_implement
}
