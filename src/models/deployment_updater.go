package models

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type DeploymentUpdater struct {
	Servicer DeploymentServicer
}

// https://godoc.org/google.golang.org/api/deploymentmanager/v2#Operation
func (u *DeploymentUpdater) Update(ctx context.Context, operation *PipelineOperation, successHandler, errorHandler UpdateHandler) error {
	newOpe, err := u.Servicer.GetOperation(ctx, operation.ProjectID, operation.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to get deployment operation: %v because of %v\n", operation, err)
		return err
	}
	oldStatus := operation.Status
	if oldStatus == newOpe.Status {
		return nil
	}

	operation.Status = newOpe.Status
	if newOpe.Status != "DONE" {
		operation.AppendLog(fmt.Sprintf("StatusChange from %s to %s", operation.Status, newOpe.Status))
		err := operation.Update(ctx)
		if err != nil {
			return err
		}
		return nil
	}

	errors := u.ErrorsFromOperation(newOpe)
	var f UpdateHandler
	if errors != nil {
		operation.Errors = *errors
		operation.AppendLog(fmt.Sprintf("Error by %v", newOpe))
		f = errorHandler
	} else {
		operation.AppendLog("Success")
		f = successHandler
	}

	err = operation.Update(ctx)
	if err != nil {
		return err
	}

	err = f(newOpe.EndTime)
	if err != nil {
		return err
	}
	return nil
}

func (u *DeploymentUpdater) ErrorsFromOperation(ope *deploymentmanager.Operation) *[]OperationError {
	doe := ope.Error
	if doe != nil && len(doe.Errors) > 0 {
		errors := []OperationError{}
		for _, e := range doe.Errors {
			errors = append(errors, OperationError{
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
