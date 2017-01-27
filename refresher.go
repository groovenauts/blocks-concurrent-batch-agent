package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Refresher struct {
	deployer DeploymentServicer
}

func (b *Refresher) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Refreshing pipeline %v\n", pl)
	if pl.Props.Status == deploying {
		b.UpdateStatusByDeployment(ctx, pl)
	}
	return nil
}

func (b *Refresher) UpdateStatusByDeployment(ctx context.Context, pl *Pipeline) error {
	// See the "Examples" below "Response"
	//   https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/insert#response
	proj := pl.Props.ProjectID
	dep_name := pl.Props.DeploymentName
	deployment, err := b.deployer.Get(ctx, proj, dep_name)
	if err != nil {
		log.Errorf(ctx, "Failed to get deployment %v\nproject: %v deployment: %v\nhc: %v\n", err, proj, dep_name)
		return err
	}
	if deployment.Operation.Status == "DONE" {
		doe := deployment.Operation.Error
		if doe != nil && len(doe.Errors) > 0 {
			errors := []DeploymentError{}
			for _, e := range doe.Errors {
				errors = append(errors, DeploymentError{
					Code: e.Code,
					Location: e.Location,
					Message: e.Message,
				})
			}
			pl.Props.Errors = errors
			pl.Props.Status = broken
		} else {
			pl.Props.Status = opened
		}
		err = pl.update(ctx)
		if err != nil {
			log.Errorf(ctx, "Failed to update Pipeline Status to %v: %v\npl: %v\n", pl.Props.Status, err, pl)
			return err
		}
	}
	return nil
}
