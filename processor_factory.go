package pipeline

import (
	"fmt"
	"golang.org/x/net/context"
)

type ProcessorFactory interface {
	Create(ctx context.Context, name string) (Processor, error)
}

type DefaultProcessorFactory struct{}

func (dpf *DefaultProcessorFactory) Create(ctx context.Context, action string) (Processor, error) {
	switch action {
	case "build":
		return &Builder{}, nil
	case "update":
		return &Updater{}, nil
	case "resize":
		return &Resizer{}, nil
	case "close":
		return &Closer{}, nil
	case "refresh":
		return &Refresher{}, nil
	default:
		return nil, fmt.Errorf("Unknown processor action: %v\n", action)
	}
}
