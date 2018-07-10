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

// PipelineBaseClosingTaskController implements the PipelineBaseClosingTask resource.
type PipelineBaseClosingTaskController struct {
	*goa.Controller
}

// NewPipelineBaseClosingTaskController creates a PipelineBaseClosingTask controller.
func NewPipelineBaseClosingTaskController(service *goa.Service) *PipelineBaseClosingTaskController {
	return &PipelineBaseClosingTaskController{Controller: service.NewController("PipelineBaseClosingTaskController")}
}

// Start runs the start action.
func (c *PipelineBaseClosingTaskController) Start(ctx *app.StartPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Start: start_implement

	// Put your logic here
	base := PipelineBaseTaskBase{
		MainStatus: model.ClosingStarting,
		NextStatus: model.ClosingRunning,
		SkipStatuses: []model.PipelineBaseStatus{
			model.ClosingRunning,
			model.ClosingError,
			model.Closed,
		},
		ProcessorFactory: func(ctx context.Context) (model.PipelineBaseProcessor, error) {
			return model.NewPipelineBaseCloser(ctx)
		},
		WatchTaskPathFunc: func(ope *model.CloudAsyncOperation) string {
			return fmt.Sprintf("/opening_tasks/%d", ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}
	return base.Start(appengine.NewContext(ctx.Request), ctx.ResourceID)

	// PipelineBaseClosingTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *PipelineBaseClosingTaskController) Watch(ctx *app.WatchPipelineBaseClosingTaskContext) error {
	// PipelineBaseClosingTaskController_Watch: start_implement

	// Put your logic here
	base := PipelineBaseTaskBase{
		MainStatus:  model.ClosingRunning,
		NextStatus:  model.Closed,
		ErrorStatus: model.ClosingError,
		SkipStatuses: []model.PipelineBaseStatus{
			model.ClosingError,
			model.Closed,
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
			return fmt.Sprintf("/closing_tasks/%d", ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondAccepted:  ctx.Accepted,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}
	return base.Watch(appengine.NewContext(ctx.Request), ctx.ID)

	// PipelineBaseClosingTaskController_Watch: end_implement
}
