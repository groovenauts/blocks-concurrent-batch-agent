package controller

import (
	"fmt"
	"time"

	"google.golang.org/appengine"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
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
	appCtx := appengine.NewContext(ctx.Request)

	t := time.Now()
	org := &model.Organization{
		Name: "dummy-" + t.Format(time.RFC3339),
		TokenAmount: 100,
	}
	key, err := org.Create(appCtx)
	if err != nil {
		return ctx.BadRequest(goa.ErrBadRequest(err))
	}

	auth := &model.Auth{ParentKey: key}
	err = auth.Create(appCtx)
	if err != nil {
		return ctx.BadRequest(goa.ErrBadRequest(err))
	}

	res := &app.DummyAuth{
		OrganizationID: fmt.Sprintf("%s", key.IntID()),
		Token: auth.Token,
	}
	return ctx.Created(res)

	// DummyAuthsController_Create: end_implement
}
