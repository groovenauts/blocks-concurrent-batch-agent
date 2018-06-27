package controller

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// InstanceGroupDestructionTaskController implements the InstanceGroupDestructionTask resource.
type InstanceGroupDestructionTaskController struct {
	*goa.Controller
}

// NewInstanceGroupDestructionTaskController creates a InstanceGroupDestructionTask controller.
func NewInstanceGroupDestructionTaskController(service *goa.Service) *InstanceGroupDestructionTaskController {
	return &InstanceGroupDestructionTaskController{Controller: service.NewController("InstanceGroupDestructionTaskController")}
}

// Start runs the start action.
func (c *InstanceGroupDestructionTaskController) Start(ctx *app.StartInstanceGroupDestructionTaskContext) error {
	// InstanceGroupDestructionTaskController_Start: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)
	store := &model.InstanceGroupStore{}
	m, err := store.Get(appCtx, ctx.ResourceID)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "InstanceGroup not found for %q\n", ctx.ResourceID)
			return ctx.NoContent(nil)
		} else {
			return err
		}
	}
	switch m.Status {
	case model.DestructionStarting: // Through
	case
		model.DestructionRunning,
		model.DestructionError,
		model.Destructed      :
		log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.OK(nil)
	default:
		log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.NoContent(nil)
	}

	return model.WithNewDestructor(appCtx, func(destructor *model.Destructor) error {
		ope, err := destructor.Process(appCtx, m)
		if err != nil {
			return err
		}
		opeStore := &model.CloudAsyncOperationStore{}
		_, err = opeStore.Put(appCtx, ope)
		if err != nil {
			return err
		}
		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			m.Status = model.DestructionRunning
			_, err = store.Put(c, m)
			if err != nil {
				return err
			}
			task := &taskqueue.Task{Method: "PUT", Path: "/destruction_tasks/" + ope.Id}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
		}, nil)
	})
	// InstanceGroupDestructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupDestructionTaskController) Watch(ctx *app.WatchInstanceGroupDestructionTaskContext) error {
	// InstanceGroupDestructionTaskController_Watch: start_implement

	// Put your logic here

	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupDestructionTaskController_Watch: end_implement
}
