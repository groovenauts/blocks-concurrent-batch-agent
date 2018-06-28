package model

import (
	"golang.org/x/net/context"
)

type PipelineBaseProcessor interface {
	Process(context.Context, *PipelineBase) (*CloudAsyncOperation, error)
}
