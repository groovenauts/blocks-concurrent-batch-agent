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
	base := InstanceGroupTaskBase{
		MainStatus: model.DestructionStarting,
		NextStatus: model.DestructionRunning,
		SkipStatuses: []model.InstanceGroupStatus{
			model.DestructionRunning,
			model.DestructionError,
			model.Destructed,
		},
		ProcessorFactory: func(ctx context.Context) (model.InstanceGroupProcessor, error) {
			return model.NewInstanceGroupDestructor(ctx)
		},
		WatchTaskPathFunc: func(ope *model.CloudAsyncOperation) string {
			return "/destruction_tasks/" + ope.Id
		},
		RespondOK: ctx.OK,
		RespondNoContent: ctx.NoContent,
		RespondCreated: ctx.Created,
	}
	return base.Start(appengine.NewContext(ctx.Request), ctx.ResourceID)

	// InstanceGroupDestructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupDestructionTaskController) Watch(ctx *app.WatchInstanceGroupDestructionTaskContext) error {
	// InstanceGroupDestructionTaskController_Watch: start_implement

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
	case model.DestructionRunning: // Through
	case
		model.DestructionError,
		model.Destructed:
		log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.OK(nil)
	default:
		log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.NoContent(nil)
	}

	servicer, err := model.DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil
	}
	remoteOpeOriginal, err := servicer.GetOperation(appCtx, ope.ProjectId, ope.Name)
	if err != nil {
		log.Errorf(appCtx, "Failed to get deployment operation: %v because of %v\n", ope, err)
		return err
	}
	remoteOpe := &model.RemoteOperationWrapperOfDeploymentmanager{
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
		if err := PutTask(appCtx, "/destruction_tasks/" + ope.Id, 1 * time.Minute); err != nil {
			return err
		}
		return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
	}

	errors := remoteOpe.Errors()
	var f func(r *app.CloudAsyncOperation) error
	if errors != nil {
		ope.Errors = *errors
		ope.AppendLog(fmt.Sprintf("Error by %v", remoteOpe.GetOriginal()))
		m.Status = model.DestructionError
		f = ctx.NoContent
	} else {
		ope.AppendLog("Success")
		m.Status = model.Destructed
		f = ctx.Accepted
	}

	_, err = opeStore.Update(appCtx, ope)
	if err != nil {
		return err
	}

	return datastore.RunInTransaction(appCtx, func(c context.Context) error {
		_, err = store.Update(c, m)
		if err != nil {
			return err
		}
		// TODO Add calling PUT /pipeline_bases/:id/hibernation_done_task
		return f(CloudAsyncOperationModelToMediaType(ope))
	}, nil)

	// InstanceGroupDestructionTaskController_Watch: end_implement
}
