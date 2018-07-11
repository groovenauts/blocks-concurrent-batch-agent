package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type InstanceGroupDestructor struct {
	deployer DeploymentServicer
}

func NewInstanceGroupDestructor(ctx context.Context) (*InstanceGroupDestructor, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &InstanceGroupDestructor{
		deployer: deployer,
	}, nil
}

func WithNewInstanceGroupDestructor(ctx context.Context, f func(*InstanceGroupDestructor) error) error {
	closer, err := NewInstanceGroupDestructor(ctx)
	if err != nil {
		return err
	}
	return f(closer)
}

func (b *InstanceGroupDestructor) Process(ctx context.Context, pl *InstanceGroup) (*InstanceGroupOperation, error) {
	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.ProjectID, pl.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, pl.Name)
		return nil, err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.ProjectID, pl.Name)

	operation := &InstanceGroupOperation{
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
