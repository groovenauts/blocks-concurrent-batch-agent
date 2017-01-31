package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type Refresher struct {
	deployer DeploymentServicer
}

func (b *Refresher) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Refreshing pipeline %v\n", pl)
	switch pl.Props.Status {
	case deploying:
		return b.UpdateDeployingPipeline(ctx, pl)
	case closing:
		return b.UpdateClosingPipeline(ctx, pl)
	default:
		return nil
	}
}

func (b *Refresher) UpdateDeployingPipeline(ctx context.Context, pl *Pipeline) error {
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
		errors := b.ErrorsFromOperation(deployment.Operation)
		if errors != nil {
			log.Errorf(ctx, "Deployment error found for project: %v deployment: %v\n%v\n", proj, dep_name, errors)
			pl.Props.DeployingErrors = *errors
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

func (b *Refresher) UpdateClosingPipeline(ctx context.Context, pl *Pipeline) error {
	proj := pl.Props.ProjectID
	ope_name := pl.Props.ClosingOperationName
	ope, err := b.deployer.GetOperation(ctx, proj, ope_name)
	if err != nil {
		log.Errorf(ctx, "Failed to get deployment project: %v deployment: %v\n%v\n", proj, ope_name, err)
		return err
	}
	log.Debugf(ctx, "Refreshing closing operation: %v\n", ope)
	if ope.Status == "DONE" {
		errors := b.ErrorsFromOperation(ope)
		if errors != nil {
			log.Errorf(ctx, "Closing error found for project: %v deployment: %v\n%v\n", proj, pl.Props.DeploymentName, errors)
			pl.Props.ClosingErrors = *errors
			pl.Props.Status = closing_error
		} else {
			log.Infof(ctx, "Closing completed successfully project: %v deployment: %v\n", proj, pl.Props.DeploymentName)
			pl.Props.Status = closed
		}
		err = pl.update(ctx)
		if err != nil {
			log.Errorf(ctx, "Failed to update Pipeline Status to %v: %v\npl: %v\n", pl.Props.Status, err, pl)
			return err
		}
	}
	return nil
}

func (b *Refresher) ErrorsFromOperation(ope *deploymentmanager.Operation) *[]DeploymentError {
	doe := ope.Error
	if doe != nil && len(doe.Errors) > 0 {
		errors := []DeploymentError{}
		for _, e := range doe.Errors {
			errors = append(errors, DeploymentError{
				Code:     e.Code,
				Location: e.Location,
				Message:  e.Message,
			})
		}
		return &errors
	} else {
		return nil
	}
}
