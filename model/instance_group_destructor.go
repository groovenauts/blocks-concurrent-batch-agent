package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Destructor struct {
	deployer DeploymentServicer
}

func NewDestructor(ctx context.Context) (*Destructor, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &Destructor{
		deployer: deployer,
	}, nil
}

func WithNewDestructor(ctx context.Context, f func(*Destructor) error) error {
	closer, err := NewDestructor(ctx)
	if err != nil {
		return err
	}
	return f(closer)
}

func (b *Destructor) Process(ctx context.Context, pl *InstanceGroup) (*CloudAsyncOperation, error) {
	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.ProjectID, pl.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, pl.Name)
		return nil, err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.ProjectID, pl.Name)

	operation := &CloudAsyncOperation{
		OwnerType:     "InstanceGroup",
		OwnerID:       pl.Id,
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
