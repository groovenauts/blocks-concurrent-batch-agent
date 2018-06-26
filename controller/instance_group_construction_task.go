package controller

import (
	"fmt"

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
	store := &model.InstanceGroupStore{}
	m, err := store.Get(c, ctx.ResourceID)
	if err != nil {
		return err
	}
	switch m.Status {
	case model.ConstructionStarting:
		b, err := model.NewConstructor(appCtx)
		if err != nil {
			return err
		}
		ope, err := b.Process(appCtx, m)
		if err != nil {
			return err
		}
		opeStore := &model.CloudAsyncOperationStore{}
		_, err = opeStore.Put(appCtx, ope)
		if err != nil {
			return err
		}
		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			m.Status = model.ConstructionRunning
			_, err = store.Put(c, m)
			if err != nil {
				return err
			}
			task := &taskqueue.Task{Method: "PUT", Path: "/construction_tasks/" + ope.Id}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
		}, nil)
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
		case model.ConstructionRunning: // Through
		case
			model.ConstructionError,
			model.Constructed:
			log.Infof(c, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return ctx.OK(nil)
		default:
			log.Warningf(c, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return ctx.NoContent(nil)
		}

		servicer, err := model.DefaultDeploymentServicer(ctx)
		if err != nil {
			return nil
		}
		remoteOpe, err := servicer.GetOperation(ctx, ope.ProjectId, ope.Name)
		if err != nil {
			log.Errorf(ctx, "Failed to get deployment operation: %v because of %v\n", ope, err)
			return err
		}
		if ope.Status != remoteOpe.Status {
			ope.AppendLog(fmt.Sprintf("InstanceGroup %q Status changed from %q to %q", m.Id, ope.Status, remoteOpe.Status))
		}

		// PENDING, RUNNING, or DONE
		switch remoteOpe.Status {
		case "DONE": // through
		default:
			if ope.Status == remoteOpe.Status {
				return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
			}
			ope.Status = remoteOpe.Status
			_, err := opeStore.Update(ctx, ope)
			if err != nil {
				return err
			}
			return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
		}

		errors := model.ErrorsFromDeploymentmanagerOperation(remoteOpe)
		var f func(r *CloudAsyncOperation) error
		if errors != nil {
			ope.Errors = *errors
			ope.AppendLog(fmt.Sprintf("Error by %v", remoteOpe))
			m.Status = ConstructionError
			f = ctx.NoContent
		} else {
			ope.AppendLog("Success")
			m.Status = model.Constructed
			f = ctx.Accepted
		}

		err := opeStore.Update(c, ope)
		if err != nil {
			return err
		}
		err = store.Update(c, m)
		if err != nil {
			return err
		}
		// TODO Add calling PUT /pipeline_bases/:id/wakeup_task
		return f(ope)
	}, nil)

	// InstanceGroupConstructionTaskController_Watch: end_implement
}
