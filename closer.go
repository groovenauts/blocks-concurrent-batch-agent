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
	_, err := b.deployer.Delete(pl.Props.ProjectID, pl.Props.Name).Context(ctx).Do()
	if err != nil {
		log.Errorf(ctx, "Failed to close deployment %v\nproject: %v deployment: %v\nhc: %v\n", err, pl.Props.ProjectID, pl.Props.Name)
		return err
	}

	pl.Props.Status = closed
	err = pl.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'closed': %v\npl: %v\n", err, pl)
		return err
	}

	return nil
}
