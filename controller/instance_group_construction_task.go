package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
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
	base := InstanceGroupTaskBase{
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
		WatchTaskPathFunc: func(ope *model.InstanceGroupOperation) string {
			return fmt.Sprintf("/construction_tasks/%d", ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}
	return base.Start(appengine.NewContext(ctx.Request), ctx.OrgID, ctx.Name)

	// InstanceGroupConstructionTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupConstructionTaskController) Watch(ctx *app.WatchInstanceGroupConstructionTaskContext) error {
	// InstanceGroupConstructionTaskController_Watch: start_implement

	// Put your logic here
	base := InstanceGroupTaskBase{
		MainStatus:  model.ConstructionRunning,
		NextStatus:  model.Constructed,
		ErrorStatus: model.ConstructionError,
		SkipStatuses: []model.InstanceGroupStatus{
			model.ConstructionError,
			model.Constructed,
		},
		RemoteOpeFunc: func(ctx context.Context, ope *model.InstanceGroupOperation) (model.RemoteOperationWrapper, error) {
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
		WatchTaskPathFunc: func(ope *model.InstanceGroupOperation) string {
			return fmt.Sprintf("/construction_tasks/%d", ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondAccepted:  ctx.Accepted,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}
	return base.Watch(appengine.NewContext(ctx.Request), ctx.OrgID, ctx.Name, ctx.ID)

	// InstanceGroupConstructionTaskController_Watch: end_implement
}
