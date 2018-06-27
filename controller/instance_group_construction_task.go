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
	start := InstanceGroupTaskStart{
		MainStatus: model.ConstructionStarting,
		NextStatus: model.ConstructionRunning,
		SkipStatuses: []model.InstanceGroupStatus{
			model.ConstructionRunning,
			model.ConstructionError,
			model.Constructed,
		},
		ProcessorFactory: func(ctx context.Context) (model.InstanceGroupProcessor, error) {
			return model.NewInstanceGroupConstructor(ctx)
		},
		WatchTaskPathFunc: func(ope *model.CloudAsyncOperation) string {
			return "/construction_tasks/" + ope.Id
		},
		RespondOk: func(ope *app.CloudAsyncOperation) error {
			return ctx.OK(ope)
		},
		RespondNoContent: func(ope *app.CloudAsyncOperation) error {
			return ctx.NoContent(ope)
		},
		RespondCreated: func(ope *app.CloudAsyncOperation) error {
			return ctx.Created(ope)
		},
	}
	return start.Run(appengine.NewContext(ctx.Request), ctx.ResourceID)

	// InstanceGroupConstructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupConstructionTaskController) Watch(ctx *app.WatchInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Watch: start_implement

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
	case model.ConstructionRunning: // Through
	case
		model.ConstructionError,
		model.Constructed:
		log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.OK(nil)
	default:
		log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
		return ctx.NoContent(nil)
	}

	servicer, err := model.DefaultDeploymentServicer(appCtx)
	if err != nil {
		return nil
	}
	remoteOpe, err := servicer.GetOperation(appCtx, ope.ProjectId, ope.Name)
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
		if err := PutTask(appCtx, "/construction_tasks/" + ope.Id, 1 * time.Minute); err != nil {
			return err
		}
		return ctx.Created(CloudAsyncOperationModelToMediaType(ope))
	}

	errors := model.ErrorsFromDeploymentmanagerOperation(remoteOpe)
	var f func(r *app.CloudAsyncOperation) error
	if errors != nil {
		ope.Errors = *errors
		ope.AppendLog(fmt.Sprintf("Error by %v", remoteOpe))
		m.Status = model.ConstructionError
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

	return datastore.RunInTransaction(appCtx, func(c context.Context) error {
		_, err = store.Update(c, m)
		if err != nil {
			return err
		}
		// TODO Add calling PUT /pipeline_bases/:id/wakeup_task
		return f(CloudAsyncOperationModelToMediaType(ope))
	}, nil)

	// InstanceGroupConstructionTaskController_Watch: end_implement
}
