package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type PipelineBaseDestructor struct {
	deployer DeploymentServicer
}

func NewPipelineBaseDestructor(ctx context.Context) (*PipelineBaseDestructor, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &PipelineBaseDestructor{
		deployer: deployer,
	}, nil
}

func WithNewPipelineBaseDestructor(ctx context.Context, f func(*PipelineBaseDestructor) error) error {
	closer, err := NewPipelineBaseDestructor(ctx)
	if err != nil {
		return err
	}
	return f(closer)
}

func (b *PipelineBaseDestructor) Process(ctx context.Context, pl *PipelineBase) (*CloudAsyncOperation, error) {
	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.InstanceGroup.ProjectID, pl.InstanceGroup.DeploymentName)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.InstanceGroup.ProjectID, pl.InstanceGroup.DeploymentName)
		return nil, err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.InstanceGroup.ProjectID, pl.InstanceGroup.DeploymentName)

	operation := &CloudAsyncOperation{
		OwnerType:     "PipelineBase",
		OwnerID:       pl.Id,
		ProjectId:     pl.InstanceGroup.ProjectID,
		Zone:          pl.InstanceGroup.Zone,
		Service:       "deploymentmanager",
		Name:          ope.Name,
		OperationType: ope.OperationType,
		Status:        ope.Status,
		Logs: []CloudAsyncOperationLog{
			CloudAsyncOperationLog{CreatedAt: time.Now(), Message: "Start"},
		},
	}

	return operation, nil
}
