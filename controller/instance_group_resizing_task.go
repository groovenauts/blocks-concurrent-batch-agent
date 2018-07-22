package controller

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine"
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
		WatchTaskPathFunc: func(ope *model.InstanceGroupOperation) string {
			return pathToInstanceGroupResizingTask(ctx.OrgID, ctx.Name, ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}

	base.Routes(
		map[model.InstanceGroupStatus]InstanceGroupTaskBaseAction{
			model.ResizeStarting: base.RunProcessorFunc(model.ResizeRunning, func(ctx context.Context) (model.InstanceGroupProcessor, error) {
				return model.NewInstanceGroupScaler(ctx)
			}),
			model.ResizeRunning: base.Skip,
		})

	return base.Start(appengine.NewContext(ctx.Request), ctx.OrgID, ctx.Name)

	// InstanceGroupResizingTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *InstanceGroupResizingTaskController) Watch(ctx *app.WatchInstanceGroupResizingTaskContext) error {
	// InstanceGroupResizingTaskController_Watch: start_implement

	// Put your logic here
	base := InstanceGroupTaskBase{
		WatchTaskPathFunc: func(ope *model.InstanceGroupOperation) string {
			return pathToInstanceGroupResizingTask(ctx.OrgID, ctx.Name, ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondAccepted:  ctx.Accepted,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}

	base.Routes(
		map[model.InstanceGroupStatus]InstanceGroupTaskBaseAction{
			model.ResizeRunning: base.SyncWithRemoteOpeFunc(model.Constructed, model.Constructed,
				func(ctx context.Context, ope *model.InstanceGroupOperation) (model.RemoteOperationWrapper, error) {
					servicer, err := model.DefaultInstanceGroupServicer(ctx)
					if err != nil {
						return nil, err
					}
					remoteOpeOriginal, err := servicer.GetZoneOp(ope.ProjectId, ope.Zone, ope.Name)
					if err != nil {
						log.Errorf(ctx, "Failed to get deployment operation: %v because of %v\n", ope, err)
						return nil, err
					}
					return &model.RemoteOperationWrapperOfCompute{
						Original: remoteOpeOriginal,
					}, nil
				},
				nil),
			model.Constructed: base.Skip,
		})

	return base.Watch(appengine.NewContext(ctx.Request), ctx.OrgID, ctx.Name, ctx.ID)

	// InstanceGroupResizingTaskController_Watch: end_implement
}
