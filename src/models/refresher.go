package models

import (
	"fmt"

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

func (b *Refresher) Process(ctx context.Context, pl *Pipeline, handler func(*[]DeploymentError) error) error {
	log.Debugf(ctx, "Refreshing pipeline %v\n", pl)

	switch pl.Status {
	case Deploying, Closing:
		b.Setup(ctx, pl)
		return b.Refresh(ctx, pl, handler)
	default:
		return nil
	}
}

func (b *Refresher) Refresh(ctx context.Context, pl *Pipeline, handler func(*[]DeploymentError) error) error {
	status := pl.Status.String()

	var ope_name string
	switch pl.Status {
	case Deploying:
		ope_name = pl.DeployingOperationName
	case Closing:
		ope_name = pl.ClosingOperationName
	default:
		return &InvalidOperation{Msg: fmt.Sprintf("Invalid Status %v to refresh Pipline %q\n", pl.Status, pl.ID)}
	}

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
		err = handler(errors)
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
