package controller

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// InstanceGroupResizingTaskController implements the InstanceGroupResizingTask resource.
type InstanceGroupResizingTaskController struct {
	*goa.Controller
}

// NewInstanceGroupResizingTaskController creates a InstanceGroupResizingTask controller.
func NewInstanceGroupResizingTaskController(service *goa.Service) *InstanceGroupResizingTaskController {
	return &InstanceGroupResizingTaskController{Controller: service.NewController("InstanceGroupResizingTaskController")}
}

// Start runs the start action.
func (c *InstanceGroupResizingTaskController) Start(ctx *app.StartInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Start: start_implement

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
	case model.ResizeStarting: // Through
	case model.ResizeRunning:
		log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.OK(nil)
	default:
		log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.NoContent(nil)
	}

	return model.WithScaler(appCtx, func(scaler *model.Scaler) error {
		ope, err := scaler.Process(appCtx, m)
		if err != nil {
			return err
		}
		opeStore := &model.CloudAsyncOperationStore{}
		_, err = opeStore.Put(appCtx, ope)
		if err != nil {
			return err
		}

		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			m.Status = model.ResizeRunning
			_, err = store.Put(c, m)
			if err != nil {
				return err
			}
			task := &taskqueue.Task{Method: "PUT", Path: "/resizing_tasks/" + ope.Id}
			if _, err := taskqueue.Add(c, task, ""); err != nil {
				return err
			}
			return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
		}, nil)
	})

	// InstanceGroupResizingTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupResizingTaskController) Watch(ctx *app.WatchInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Watch: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)
	opeStore := &model.CloudAsyncOperationStore{}
	ope, err := opeStore.Get(appCtx, ctx.ID)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "CloudAsyncOperation not found for %q\n", ctx.ID)
			return ctx.NoContent(nil)
		} else {
			return err
		}
	}
	store := &model.InstanceGroupStore{}
	m, err := store.Get(appCtx, ope.OwnerID)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "InstanceGroup not found for %q\n", ope.OwnerID)
			return ctx.NoContent(nil)
		} else {
			return err
		}
	}

	switch m.Status {
	case model.ResizeRunning: // Through
	default:
		log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.NoContent(nil)
	}

	servicer, err := model.DefaultInstanceGroupServicer(ctx)
	if err != nil {
		return nil
	}
	remoteOpe, err := servicer.GetZoneOp(ope.ProjectId, ope.Zone, ope.Name)
	if err != nil {
		log.Errorf(appCtx, "Failed to get deployment operation: %v because of %v\n", ope, err)
		return err
	}
	if ope.Status != remoteOpe.Status {
		ope.AppendLog(fmt.Sprintf("InstanceGroup %q Status changed from %q to %q", m.Id, ope.Status, remoteOpe.Status))
	}

	// PENDING, RUNNING, or DONE
	switch remoteOpe.Status {
	case "DONE": // through
	default:
		if ope.Status != remoteOpe.Status {
			ope.Status = remoteOpe.Status
			_, err := opeStore.Update(appCtx, ope)
			if err != nil {
				return err
			}
		}
		if err := PutTask(appCtx, "/resizing_tasks/" + ope.Id, 1 * time.Minute); err != nil {
			return err
		}
		return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
	}

	errors := model.ErrorsFromComputeOperation(remoteOpe)
	var f func(r *app.CloudAsyncOperation) error
	if errors != nil {
		ope.Errors = *errors
		ope.AppendLog(fmt.Sprintf("Error by %v", remoteOpe))
		m.Status = model.Constructed
		log.Errorf(appCtx, "Failed to resize InstanceGroup %q\n", m.Id)
		f = ctx.NoContent
	} else {
		ope.AppendLog("Success")
		m.Status = model.Constructed
		f = ctx.Accepted
	}

	_, err = opeStore.Update(appCtx, ope)
	if err != nil {
		return err
	}

	_, err = store.Update(appCtx, m)
	if err != nil {
		return err
	}
	return f(CloudAsyncOperationModelToMediaType(ope))

	// InstanceGroupResizingTaskController_Watch: end_implement
}
