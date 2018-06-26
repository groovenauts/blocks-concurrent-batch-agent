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
			b, err := model.NewConstructor(c)
			if err != nil {
				return err
			}
			ope, err := b.Process(c, m)
			if err != nil {
				return err
			}
			m.Status = model.ConstructionRunning
			_, err = store.Put(c, m)
			if err != nil {
				return err
			}
			opeStore := &model.CloudAsyncOperationStore{}
			_, err = opeStore.Put(c, ope)
			if err != nil {
				return err
			}
			task := &taskqueue.Task{Method: "PUT", Path: "/construction_tasks/" + ope.Id}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
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

	}, nil)

	// InstanceGroupConstructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupConstructionTaskController) Watch(ctx *app.WatchInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Watch: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)
	return datastore.RunInTransaction(appCtx, func(c context.Context) error {
		opeStore := &model.CloudAsyncOperationStore{}
		ope, err := opeStore.Get(c, ctx.ID)
		if err != nil {
			return err
		}
		store := &model.InstanceGroupStore{}
		m, err := store.Get(c, ope.OwnerID)
		if err != nil {
			return err
		}

		switch m.Status {
		case model.ConstructionRunning:
			servicer, err := model.DefaultDeploymentServicer(ctx)
			if err != nil {
				return nil
			}
			remoteOpe, err := servicer.GetOperation(ctx, ope.ProjectId, ope.Name)
			if err != nil {
				log.Errorf(ctx, "Failed to get deployment operation: %v because of %v\n", ope, err)
				return err
			}
			switch remoteOpe.Status {
			case "DONE":
				// TODO
				return nil
			default:
				if ope.Status == remoteOpe.Status {
					return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
				}
				// ope.AppendLog(fmt.Sprintf("StatusChange from %s to %s", operation.Status, newOpe.Status))
				ope.Status = remoteOpe.Status
				_, err := opeStore.Update(ctx, ope)
				if err != nil {
					return err
				}
				return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
			}
		case
			model.ConstructionError,
			model.Constructed:
			log.Infof(c, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return ctx.OK(nil)
		default:
			log.Warningf(c, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return ctx.NoContent(nil)
		}
	}, nil)
	res := &app.CloudAsyncOperation{}
	return ctx.OK(res)
	// InstanceGroupConstructionTaskController_Watch: end_implement
}
