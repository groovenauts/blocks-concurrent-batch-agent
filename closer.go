package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/deploymentmanager/v2"
)

type Closer struct {
}

func (b *Closer) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Closing pipeline %v\n", pl.Props)

	// See the "Examples" below "Response"
	//   https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/delete
	hc, err := google.DefaultClient(ctx, deploymentmanager.CloudPlatformScope)
	if err != nil {
		log.Errorf(ctx, "Failed to get google.DefaultClient: %v\n", err)
		return err
	}
	c, err := deploymentmanager.New(hc)
	if err != nil {
		log.Errorf(ctx, "Failed to get deploymentmanager.New(hc): %v\nhc: %v\n", err, hc)
		return err
	}
	_, err = c.Deployments.Delete(pl.Props.ProjectID, pl.Props.Name).Context(ctx).Do()
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
