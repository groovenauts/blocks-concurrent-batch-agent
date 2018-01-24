package models

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Closer struct {
	deployer DeploymentServicer
}

func NewCloser(ctx context.Context) (*Closer, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &Closer{
		deployer: deployer,
	}, nil
}

func (b *Closer) Process(ctx context.Context, pl *Pipeline) (*PipelineOperation, error) {
	log.Debugf(ctx, "Closing pipeline %v\n", pl)

	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.ProjectID, pl.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, pl.Name)
		return nil, err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.ProjectID, pl.Name)

	operation := &PipelineOperation{
		Pipeline:      pl,
		ProjectID:     pl.ProjectID,
		Zone:          pl.Zone,
		Service:       "deploymentmanager",
		Name:          ope.Name,
		OperationType: ope.OperationType,
		Status:        ope.Status,
		Logs: []OperationLog{
			OperationLog{CreatedAt: time.Now(), Message: "Start"},
		},
	}
	err = operation.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create PipelineOperation: %v because of %v\n", operation, err)
		return nil, err
	}

	return operation, nil
}
