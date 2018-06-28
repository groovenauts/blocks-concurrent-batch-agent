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

type RemoteOperationWrapperOfDeploymentmanager struct {
	Original *deploymentmanager.Operation
}

func (w *RemoteOperationWrapperOfDeploymentmanager) GetOriginal() interface{} {
	return w.Original
}

func (w *RemoteOperationWrapperOfDeploymentmanager) Status() string {
	return w.Original.Status
}

func (w *RemoteOperationWrapperOfDeploymentmanager) Errors() *[]CloudAsyncOperationError {
	return ErrorsFromDeploymentmanagerOperation(w.Original)
}

type (
	Resource struct {
		Type       string                 `json:"type"`
		Name       string                 `json:"name"`
		Properties map[string]interface{} `json:"properties"`
	}

	Resources struct {
		Resources []Resource `json:"resources"`
	}
)
