package model

import (
	"google.golang.org/api/deploymentmanager/v2"
)

func ErrorsFromDeploymentmanagerOperation(ope *deploymentmanager.Operation) *[]CloudAsyncOperationError {
	doe := ope.Error
	if doe != nil && len(doe.Errors) > 0 {
		errors := []CloudAsyncOperationError{}
		for _, e := range doe.Errors {
			errors = append(errors, CloudAsyncOperationError{
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
