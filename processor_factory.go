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
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	switch action {
	case "build":
		return &Builder{deployer: deployer}, nil
	case "close":
		return &Closer{deployer: deployer}, nil
	case "refresh":
		return &Refresher{deployer: deployer}, nil
	default:
		return nil, fmt.Errorf("Unknown processor action: %v\n", action)
	}
}
