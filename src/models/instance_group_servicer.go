package models

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/appengine/log"
)

type InstanceGroupServicer interface {
	GetIg(project, zone, instanceGroup string) (*compute.InstanceGroup, error)
	Resize(project, zone, instanceGroupManager string, size int64) (*compute.Operation, error)
	GetZoneOp(project, zone, operation string) (*compute.Operation, error)
}

func DefaultInstanceGroupServicer(ctx context.Context) (InstanceGroupServicer, error) {
	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/insert#examples
	hc, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Errorf(ctx, "Failed to get google.DefaultClient: %v\n", err)
		return nil, err
	}
	c, err := compute.New(hc)
	if err != nil {
		log.Errorf(ctx, "Failed to get compute.New(hc): %v\nhc: %v\n", err, hc)
		return nil, err
	}
	return &InstanceGroupServiceWrapper{
		igService:      c.InstanceGroups,
		igmService:     c.InstanceGroupManagers,
		zoneOpsService: c.ZoneOperations,
	}, nil
}

func WithInstanceGroupServicer(ctx context.Context, f func(InstanceGroupServicer) error) error {
	servicer, err := DefaultInstanceGroupServicer(ctx)
	if err != nil {
		return err
	}
	return f(servicer)
}

type InstanceGroupServiceWrapper struct {
	igService      *compute.InstanceGroupsService
	igmService     *compute.InstanceGroupManagersService
	zoneOpsService *compute.ZoneOperationsService
}

func (w *InstanceGroupServiceWrapper) GetIg(project, zone, instanceGroup string) (*compute.InstanceGroup, error) {
	return w.igService.Get(project, zone, instanceGroup).Do()
}

func (w *InstanceGroupServiceWrapper) Resize(project, zone, instanceGroupManager string, size int64) (*compute.Operation, error) {
	return w.igmService.Resize(project, zone, instanceGroupManager, size).Do()
}

func (w *InstanceGroupServiceWrapper) GetZoneOp(project, zone, operation string) (*compute.Operation, error) {
	return w.zoneOpsService.Get(project, zone, operation).Do()
}
