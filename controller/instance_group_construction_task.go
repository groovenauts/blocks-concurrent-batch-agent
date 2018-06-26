package controller

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// InstanceGroupConstructionTaskController implements the InstanceGroupConstructionTask resource.
type InstanceGroupConstructionTaskController struct {
	*goa.Controller
}

// NewInstanceGroupConstructionTaskController creates a InstanceGroupConstructionTask controller.
func NewInstanceGroupConstructionTaskController(service *goa.Service) *InstanceGroupConstructionTaskController {
	return &InstanceGroupConstructionTaskController{Controller: service.NewController("InstanceGroupConstructionTaskController")}
}

// Start runs the start action.
func (c *InstanceGroupConstructionTaskController) Start(ctx *app.StartInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Start: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)
	return datastore.RunInTransaction(appCtx, func(c context.Context) error {
		store := &model.InstanceGroupStore{}
		m, err := store.Get(c, ctx.ResourceID)
		if err != nil {
			return err
		}
		switch m.Status {
		case model.ConstructionStarting:
			b, err := model.NewBuilder(c)
			if err != nil {
				return err
			}
			ope, err := b.Process(c, m)
			if err != nil {
				return err
			}
			err = store.Put(c, m)
			if err != nil {
				return err
			}
			opeStore := &model.CloudAsyncOperationStore{}
			err = opeStore.Put(c, ope)
			if err != nil {
				return err
			}
			task := &taskqueue.Task{Method: "PUT", Path: "/construction_tasks/" + ope.Name}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return nil
		case
			model.ConstructionRunning,
			model.ConstructionError,
			model.Constructed:
			log.Infof(c, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return ctx.OK(nil)
		default:
			log.Warningf(c, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return ctx.NoContent(nil)
		}

	})

	// InstanceGroupConstructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupConstructionTaskController) Watch(ctx *app.WatchInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupConstructionTaskController_Watch: end_implement
}
