package controller

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine"
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
	base := InstanceGroupTaskBase{
		MainStatus: model.DestructionRunning,
		NextStatus: model.Destructed,
		ErrorStatus: model.DestructionError,
		SkipStatuses: []model.InstanceGroupStatus{
			model.DestructionError,
			model.Destructed,
		},
		RemoteOpeFunc: func(ctx context.Context, ope *model.CloudAsyncOperation) (model.RemoteOperationWrapper, error) {
			servicer, err := model.DefaultDeploymentServicer(ctx)
			if err != nil {
				return nil, err
			}
			remoteOpeOriginal, err := servicer.GetOperation(ctx, ope.ProjectId, ope.Name)
			if err != nil {
				log.Errorf(ctx, "Failed to get deployment operation: %v because of %v\n", ope, err)
				return nil, err
			}
			return &model.RemoteOperationWrapperOfDeploymentmanager{
				Original: remoteOpeOriginal,
			}, nil
		},
		WatchTaskPathFunc: func(ope *model.CloudAsyncOperation) string {
			return "/destruction_tasks/" + ope.Id
		},
		RespondOK: ctx.OK,
		RespondAccepted: ctx.Accepted,
		RespondNoContent: ctx.NoContent,
		RespondCreated: ctx.Created,
	}
	return base.Watch(appengine.NewContext(ctx.Request), ctx.ID)

	// InstanceGroupDestructionTaskController_Watch: end_implement
}
