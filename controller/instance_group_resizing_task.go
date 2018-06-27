package controller

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

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
	base := InstanceGroupTaskBase{
		MainStatus: model.ResizeStarting,
		NextStatus: model.ResizeRunning,
		SkipStatuses: []model.InstanceGroupStatus{
			model.ResizeRunning,
		},
		ProcessorFactory: func(ctx context.Context) (model.InstanceGroupProcessor, error) {
			return model.NewInstanceGroupScaler(ctx)
		},
		WatchTaskPathFunc: func(ope *model.CloudAsyncOperation) string {
			return "/resizing_tasks/" + ope.Id
		},
		RespondOK: ctx.OK,
		RespondNoContent: ctx.NoContent,
		RespondCreated: ctx.Created,
	}
	return base.Start(appengine.NewContext(ctx.Request), ctx.ResourceID)

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
	remoteOpeOriginal, err := servicer.GetZoneOp(ope.ProjectId, ope.Zone, ope.Name)
	if err != nil {
		log.Errorf(appCtx, "Failed to get deployment operation: %v because of %v\n", ope, err)
		return err
	}
	remoteOpe := &model.RemoteOperationWrapperOfCompute{
		Original: remoteOpeOriginal,
	}

	if ope.Status != remoteOpe.Status() {
		ope.AppendLog(fmt.Sprintf("InstanceGroup %q Status changed from %q to %q", m.Id, ope.Status, remoteOpe.Status()))
	}

	// PENDING, RUNNING, or DONE
	switch remoteOpe.Status() {
	case "DONE": // through
	default:
		if ope.Status != remoteOpe.Status() {
			ope.Status = remoteOpe.Status()
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

	errors := remoteOpe.Errors()
	var f func(r *app.CloudAsyncOperation) error
	if errors != nil {
		ope.Errors = *errors
		ope.AppendLog(fmt.Sprintf("Error by %v", remoteOpe.GetOriginal()))
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
