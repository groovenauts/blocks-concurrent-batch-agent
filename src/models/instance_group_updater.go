package models

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/appengine/log"
)

type InstanceGroupUpdater struct {
	Servicer InstanceGroupServicer
}

// https://godoc.org/google.golang.org/api/compute/v1#Operation
func (u *InstanceGroupUpdater) Update(ctx context.Context, operation *PipelineOperation, successHandler, errorHandler UpdateHandler) error {
	newOpe, err := u.Servicer.GetZoneOp(operation.ProjectID, operation.Zone, operation.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to get compute operation: %v because of %v\n", operation, err)
		return err
	}
	oldStatus := operation.Status
	if oldStatus == newOpe.Status {
		return nil
	}

	operation.Status = newOpe.Status
	if newOpe.Status != "DONE" {
		operation.AppendLog(fmt.Sprintf("StatusChange from %s to %s", oldStatus, newOpe.Status))
		err := operation.Update(ctx)
		if err != nil {
			return err
		}
		return nil
	}

	errors := u.ErrorsFromOperation(newOpe)
	var f UpdateHandler
	if errors != nil {
		log.Warningf(ctx, "The operation failed: %v because of %v\n", newOpe, err)
		operation.Errors = *errors
		operation.AppendLog(fmt.Sprintf("Error by %v", newOpe))
		f = errorHandler
	} else {
		log.Infof(ctx, "The operation succeeded: %v\n", newOpe)
		operation.AppendLog("Success")
		f = successHandler
	}

	err = operation.Update(ctx)
	if err != nil {
		return err
	}

	err = f()
	if err != nil {
		return err
	}
	return nil
}

func (u *InstanceGroupUpdater) ErrorsFromOperation(ope *compute.Operation) *[]OperationError {
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
