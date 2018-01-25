package models

import (
	"golang.org/x/net/context"
)

type UpdateHandler func() error

type Updater interface {
	Update(ctx context.Context, operation *PipelineOperation, successHandler, errorHandler UpdateHandler) error
}
