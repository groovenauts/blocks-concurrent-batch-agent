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

// PipelineBaseOpeningTaskController implements the PipelineBaseOpeningTask resource.
type PipelineBaseOpeningTaskController struct {
	*goa.Controller
}

// NewPipelineBaseOpeningTaskController creates a PipelineBaseOpeningTask controller.
func NewPipelineBaseOpeningTaskController(service *goa.Service) *PipelineBaseOpeningTaskController {
	return &PipelineBaseOpeningTaskController{Controller: service.NewController("PipelineBaseOpeningTaskController")}
}

// Start runs the start action.
func (c *PipelineBaseOpeningTaskController) Start(ctx *app.StartPipelineBaseOpeningTaskContext) error {
	// PipelineBaseOpeningTaskController_Start: start_implement

	// Put your logic here
	base := PipelineBaseTaskBase{
		MainStatus: model.OpeningStarting,
		NextStatus: model.OpeningRunning,
		SkipStatuses: []model.PipelineBaseStatus{
			model.OpeningRunning,
			model.OpeningError,
			model.Hibernating,
		},
		ProcessorFactory: func(ctx context.Context) (model.PipelineBaseProcessor, error) {
			return model.NewPipelineBaseOpener(ctx)
		},
		WatchTaskPathFunc: func(ope *model.PipelineBaseOperation) string {
			return fmt.Sprintf("/opening_tasks/%d", ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}
	return base.Start(appengine.NewContext(ctx.Request), ctx.Name)

	// PipelineBaseOpeningTaskController_Start: end_implement
}

// Watch runs the watch action.
func (c *PipelineBaseOpeningTaskController) Watch(ctx *app.WatchPipelineBaseOpeningTaskContext) error {
	// PipelineBaseOpeningTaskController_Watch: start_implement

	// Put your logic here
	base := PipelineBaseTaskBase{
		MainStatus:  model.OpeningRunning,
		NextStatus:  model.Hibernating,
		ErrorStatus: model.OpeningError,
		SkipStatuses: []model.PipelineBaseStatus{
			model.OpeningError,
			model.Hibernating,
		},
		RemoteOpeFunc: func(ctx context.Context, ope *model.PipelineBaseOperation) (model.RemoteOperationWrapper, error) {
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
		WatchTaskPathFunc: func(ope *model.PipelineBaseOperation) string {
			return fmt.Sprintf("/opening_tasks/%d", ope.Id)
		},
		RespondOK:        ctx.OK,
		RespondAccepted:  ctx.Accepted,
		RespondNoContent: ctx.NoContent,
		RespondCreated:   ctx.Created,
	}
	return base.Watch(appengine.NewContext(ctx.Request), ctx.ID)

	// PipelineBaseOpeningTaskController_Watch: end_implement
}
