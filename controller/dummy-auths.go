package controller

import (
	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

// DummyAuthsController implements the dummy-auths resource.
type DummyAuthsController struct {
	*goa.Controller
}

// NewDummyAuthsController creates a dummy-auths controller.
func NewDummyAuthsController(service *goa.Service) *DummyAuthsController {
	return &DummyAuthsController{Controller: service.NewController("DummyAuthsController")}
}

// Create runs the create action.
func (c *DummyAuthsController) Create(ctx *app.CreateDummyAuthsContext) error {
	// DummyAuthsController_Create: start_implement

	// Put your logic here

	return nil
	// DummyAuthsController_Create: end_implement
}
