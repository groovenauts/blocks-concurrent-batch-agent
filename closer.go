package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Closer struct {
	deployer DeploymentServicer
}

func (b *Closer) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Closing pipeline %v\n", pl.Props)

	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete#examples
	ope, err := b.deployer.Delete(ctx, pl.Props.ProjectID, pl.Props.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\n", err, pl.Props.ProjectID, pl.Props.Name)
		return err
	}

	pl.Props.Status = closing
	pl.Props.ClosingOperationName = ope.Name
	err = pl.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'closing': %v\npl: %v\n", err, pl)
		return err
	}

	log.Infof(ctx, "Closing operation successfully started: %v deployment: %v\n", pl.Props.ProjectID, pl.Props.Name)

	return nil
}
