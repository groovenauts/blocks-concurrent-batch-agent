package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type PipelineBaseCloser struct {
	deployer DeploymentServicer
}

func NewPipelineBaseCloser(ctx context.Context) (*PipelineBaseCloser, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &PipelineBaseCloser{
		deployer: deployer,
	}, nil
}

func WithNewPipelineBaseCloser(ctx context.Context, f func(*PipelineBaseCloser) error) error {
	closer, err := NewPipelineBaseCloser(ctx)
	if err != nil {
		return err
	}
	return f(closer)
}

func (b *PipelineBaseCloser) Process(ctx context.Context, pl *PipelineBase) (*PipelineBaseOperation, error) {
	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.ProjectID, pl.InstanceGroup.DeploymentName)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, pl.InstanceGroup.DeploymentName)
		return nil, err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.ProjectID, pl.InstanceGroup.DeploymentName)

	operation := &PipelineBaseOperation{
		ProjectId:     pl.ProjectID,
		Zone:          pl.Zone,
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
