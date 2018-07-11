package model

import (
	"golang.org/x/net/context"
)

type InstanceGroupProcessor interface {
	Process(context.Context, *InstanceGroup) (*InstanceGroupOperation, error)
}
