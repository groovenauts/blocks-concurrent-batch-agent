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
		log.Errorf(ctx, "Failed to get deployment project: %v deployment: %v\n%v\n", proj, dep_name, err)
		return err
	}
	log.Debugf(ctx, "Refreshing deployment: %v\n", deployment)
	if deployment.Operation == nil {
		log.Warningf(ctx, "Deployment operation was nil for %v\nproject: %v deployment: %v\n", proj, dep_name)
		return nil
	}
	log.Debugf(ctx, "Refreshing deployment operation: %v\n", deployment.Operation)
	if deployment.Operation.Status == "DONE" {
		doe := deployment.Operation.Error
		if doe != nil && len(doe.Errors) > 0 {
			errors := []DeploymentError{}
			for _, e := range doe.Errors {
				errors = append(errors, DeploymentError{
					Code:     e.Code,
					Location: e.Location,
					Message:  e.Message,
				})
			}
			log.Errorf(ctx, "Deployment error found for project: %v deployment: %v\n%v\n", proj, dep_name, errors)
			pl.Props.Errors = errors
			pl.Props.Status = broken
		} else {
			log.Infof(ctx, "Deployment completed successfully %v\n", dep_name)
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
