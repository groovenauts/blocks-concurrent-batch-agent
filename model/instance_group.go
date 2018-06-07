package model

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/mjibson/goon"
)

type PipelineVmDisk struct {
	// DiskName    string
	DiskSizeGb  int
	DiskType    string
	SourceImage string
}

type Accelerators struct {
	Count int
	Type  string
}

type InstanceGroup struct {
	Id               string `datastore:"-" goon:"id"`
	Name             string
	ProjectID        string
	Zone             string
	BootDisk         PipelineVmDisk
	MachineType      string
	GpuAccelerators  Accelerators
	Preemptible      bool
	InstanceSize     int
	StartupScript    string
	Status           string
	DeploymentName   string
	TokenConsumption int
}

type InstanceGroupStore struct{}

func (s *InstanceGroupStore) GetAll(ctx context.Context) ([]*InstanceGroup, error) {
	g := goon.FromContext(ctx)
	r := []*InstanceGroup{}
	_, err := g.GetAll(datastore.NewQuery(g.Kind(new(InstanceGroup))), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll InstanceGroup because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *InstanceGroupStore) Put(ctx context.Context, m *InstanceGroup) (*datastore.Key, error) {
	g := goon.FromContext(ctx)
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}
