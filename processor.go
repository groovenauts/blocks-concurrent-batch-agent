package pipeline

import (
	"golang.org/x/net/context"
)

type Processor interface {
	Process(ctx context.Context, pl *Pipeline) error
}
