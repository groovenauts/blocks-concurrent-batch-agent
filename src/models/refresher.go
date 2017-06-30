package models

import (
	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type Refresher struct {
	deployer DeploymentServicer
}

func (b *Refresher) Setup(ctx context.Context, pl *Pipeline) error {
	if b.deployer == nil {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return err
	}
	b.deployer = deployer
	}

	err := pl.LoadOrganization(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to load Organization for Pipeline: %v\npl: %v\n", err, pl)
		return err
	}

	return nil
}

func (b *Refresher) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Refreshing pipeline %v\n", pl)

	switch pl.Status {
	case Deploying:
		b.Setup(ctx, pl)
		return b.UpdatePipelineWithStatus(ctx, pl, "deploying", pl.DeployingOperationName,
			func(errors *[]DeploymentError, status string) error {
				if errors != nil {
					return pl.FailDeploying(ctx, errors)
				} else {
					return pl.CompleteDeploying(ctx)
				}
			},
		)
	case Closing:
		b.Setup(ctx, pl)
		return b.UpdatePipelineWithStatus(ctx, pl, "closing", pl.ClosingOperationName,
			func(errors *[]DeploymentError, status string) error {
				if errors != nil {
					return pl.FailDeploying(ctx, errors)
				} else {
					return pl.CompleteClosing(ctx)
				}
			},
		)
	default:
		return nil
	}
}

func (b *Refresher) UpdatePipelineWithStatus(ctx context.Context, pl *Pipeline, status, ope_name string, handler func(*[]DeploymentError, string) error) error {
	// See the "Examples" below "Response"
	//   https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/insert#response
	ope, err := b.deployer.GetOperation(ctx, pl.ProjectID, ope_name)
	if err != nil {
		log.Errorf(ctx, "Failed to get %v operation project: %v deployment: %v\n%v\n", status, pl.ProjectID, ope_name, err)
		return err
	}
	log.Debugf(ctx, "Refreshing %v operation: %v\n", status, ope)
	if ope.Status == "DONE" {
		errors := b.ErrorsFromOperation(ope)
		if errors != nil {
			log.Errorf(ctx, "%v error found for project: %v deployment: %v\n%v\n", status, pl.ProjectID, pl.DeploymentName, errors)
		} else {
			log.Infof(ctx, "%v completed successfully project: %v deployment: %v\n", status, pl.ProjectID, pl.DeploymentName)
		}
		err = handler(errors, status)
		if err != nil {
			log.Errorf(ctx, "Failed to update Pipeline Status to %v: %v\npl: %v\n", pl.Status, err, pl)
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
