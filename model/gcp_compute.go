package model

import (
	"google.golang.org/api/compute/v1"
)

func ErrorsFromComputeOperation(ope *compute.Operation) *[]CloudAsyncOperationError {
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

type RemoteOperationWrapperOfCompute struct {
	Original *compute.Operation
}

func (w *RemoteOperationWrapperOfCompute) GetOriginal() interface{} {
	return w.Original
}

func (w *RemoteOperationWrapperOfCompute) Status() string {
	return w.Original.Status
}

func (w *RemoteOperationWrapperOfCompute) Errors() *[]CloudAsyncOperationError {
	return ErrorsFromComputeOperation(w.Original)
}
