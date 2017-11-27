package models

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Closer struct {
	deployer DeploymentServicer
	Handler  func(context.Context, string) error
}

func NewCloser(ctx context.Context, handler func(context.Context, string) error) (*Closer, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &Closer{
		deployer: deployer,
		Handler:  handler,
	}, nil
}

func (b *Closer) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Closing pipeline %v\n", pl)

	err := pl.LoadOrganization(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to load Organization for Pipeline: %v\npl: %v\n", err, pl)
		return err
	}

	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.ProjectID, pl.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, pl.Name)
		return err
	}

	err = b.Handler(ctx, ope.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'closing': %v\npl: %v\n", err, pl)
		return err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.ProjectID, pl.Name)

	return nil
}
