package model

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/mjibson/goon"

	"github.com/groovenauts/blocks-concurrent-batch-server/app"
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
	StackdriverAgent bool
	InstanceSize     int
	StartupScript    string
	Status           string
	DeploymentName   string
	TokenConsumption int
}

type InstanceGroupStore struct {
	ctx context.Context
}

func (s *InstanceGroupStore) GetAll() ([]*InstanceGroup, error) {
	g := goon.FromContext(ctx)
	r := []*InstanceGroup{}
	_, err := g.GetAll(datastore.NewQuery(g.Kind(new(InstanceGroup))), &r)
	if err != nil {
		log.Errorf(s.ctx, "Failed to GetAll InstanceGroup because of %v\n", err)
		return nil, err
	}
	return r
}

func (s *InstanceGroupStore) Put(m *InstanceGroup) (*datastore.Key, error) {
	g := goon.FromContext(s.ctx)
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(s.ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}
