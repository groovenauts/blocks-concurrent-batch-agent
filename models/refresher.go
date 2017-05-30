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
	switch pl.Status {
	case deploying:
		return b.UpdatePipelineWithStatus(ctx, pl, "deploying", pl.DeployingOperationName,
			func(errors *[]DeploymentError) {
				pl.DeployingErrors = *errors
				pl.Status = broken
			},
			func() {
				pl.Status = opened
			},
		)
	case closing:
		return b.UpdatePipelineWithStatus(ctx, pl, "closing", pl.ClosingOperationName,
			func(errors *[]DeploymentError) {
				pl.ClosingErrors = *errors
				pl.Status = closing_error
			},
			func() {
				pl.Status = closed
			},
		)
	default:
		return nil
	}
}

func (b *Refresher) UpdatePipelineWithStatus(ctx context.Context, pl *Pipeline, status, ope_name string, errorHandler func(*[]DeploymentError), succHandler func()) error {
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
		if errors != nil {
			log.Errorf(ctx, "%v error found for project: %v deployment: %v\n%v\n", status, proj, pl.DeploymentName, errors)
			errorHandler(errors)
		} else {
			log.Infof(ctx, "%v completed successfully project: %v deployment: %v\n", status, proj, pl.DeploymentName)
			succHandler()
		}
		err = pl.update(ctx)
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
