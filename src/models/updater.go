package models

import (
	"context"
)

// endTime is RFC3339 time string from Operation.
// https://godoc.org/google.golang.org/api/compute/v1#Operation
type UpdateHandler func(endTime string) error

type Updater interface {
	Update(ctx context.Context, operation *PipelineOperation, successHandler, errorHandler UpdateHandler) error
}
