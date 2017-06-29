package models

import (
	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type Refresher struct {
	deployer DeploymentServicer
}

func NewRefresher(ctx context.Context) (*Refresher, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &Refresher{deployer: deployer}, nil
}

func (b *Refresher) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Refreshing pipeline %v\n", pl)
	err := pl.LoadOrganization(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to load Organization for Pipeline: %v\npl: %v\n", err, pl)
		return err
	}

	proj := pl.ProjectID
	switch pl.Status {
	case Deploying:
		return b.UpdatePipelineWithStatus(ctx, pl, "deploying", pl.DeployingOperationName,
			func(errors *[]DeploymentError, status string) error {
				if errors != nil {
					log.Errorf(ctx, "%v error found for project: %v deployment: %v\n%v\n", status, proj, pl.DeploymentName, errors)
					pl.DeployingErrors = *errors
					pl.Status = Broken
				} else {
					log.Infof(ctx, "%v completed successfully project: %v deployment: %v\n", status, proj, pl.DeploymentName)
					pl.Status = Opened
				}
				return pl.Update(ctx)
			},
		)
	case Closing:
		return b.UpdatePipelineWithStatus(ctx, pl, "closing", pl.ClosingOperationName,
			func(errors *[]DeploymentError, status string) error {
				if errors != nil {
					log.Errorf(ctx, "%v error found for project: %v deployment: %v\n%v\n", status, proj, pl.DeploymentName, errors)
					pl.ClosingErrors = *errors
					pl.Status = Closing_error
					return pl.Update(ctx)
				} else {
					log.Infof(ctx, "%v completed successfully project: %v deployment: %v\n", status, proj, pl.DeploymentName)
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
	proj := pl.ProjectID
	ope, err := b.deployer.GetOperation(ctx, proj, ope_name)
	if err != nil {
		log.Errorf(ctx, "Failed to get %v operation project: %v deployment: %v\n%v\n", status, proj, ope_name, err)
		return err
	}
	log.Debugf(ctx, "Refreshing %v operation: %v\n", status, ope)
	if ope.Status == "DONE" {
		errors := b.ErrorsFromOperation(ope)
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
