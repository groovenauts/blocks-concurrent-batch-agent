package pipeline

import (
	"fmt"
	"golang.org/x/net/context"
)

type ProcessorFactory interface {
	Create(ctx context.Context, name string) (*Processor, error)
}

type DefaultProcessorFactory struct {}

func (dpf *DefaultProcessorFactory) Create(ctx context.Context, name string) (Processor, error) {
	switch name {
	case "builder": return &Builder{}, nil
	case "updater": return &Updater{}, nil
	case "resizer": return &Resizer{}, nil
	case "closer": return &Closer{}, nil
	case "refresher": return &Refresher{}, nil
	default: return nil, fmt.Errorf("Unknown processor name: %v\n", name)
	}
}
