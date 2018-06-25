package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/goadesign/goa/uuid"

	"github.com/mjibson/goon"
)

type InstanceGroupStatus string

const (
	ConstructionStarting InstanceGroupStatus = "construction_starting"
	ConstructionRunning  InstanceGroupStatus = "construction_running"
	ConstructionError    InstanceGroupStatus = "construction_error"
	Constructed          InstanceGroupStatus = "constructed"
	ResizeStarting       InstanceGroupStatus = "resize_starting"
	ResizeRunning        InstanceGroupStatus = "resize_running"
	DestructionStarting  InstanceGroupStatus = "destruction_starting"
	DestructionRunning   InstanceGroupStatus = "destruction_running"
	DestructionError     InstanceGroupStatus = "destruction_error"
	Destructed           InstanceGroupStatus = "destructed"
)

type InstanceGroupVMDisk struct {
	DiskSizeGb  int
	DiskType    string
	SourceImage string
}

type InstanceGroupAccelerators struct {
	Count int
	Type  string
}

type InstanceGroupBody struct {
	ProjectID             string
	Zone                  string
	BootDisk              InstanceGroupVMDisk
	MachineType           string
	GpuAccelerators       InstanceGroupAccelerators
	Preemptible           bool
	InstanceSizeRequested int
	InstanceSize          int
	StartupScript         string
	Status                InstanceGroupStatus
	DeploymentName        string
	TokenConsumption      int
}

type InstanceGroup struct {
	Id                    string         `datastore:"-" goon:"id"`
	Parent                *datastore.Key `datastore:"-" goon:"parent"`
	Name                  string
	ProjectID             string
	Zone                  string
	BootDisk              InstanceGroupVMDisk
	MachineType           string
	GpuAccelerators       InstanceGroupAccelerators
	Preemptible           bool
	InstanceSizeRequested int
	InstanceSize          int
	StartupScript         string
	Status                InstanceGroupStatus
	DeploymentName        string
	TokenConsumption      int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type InstanceGroupStore struct {
	ParentKey *datastore.Key
}

func (s *InstanceGroupStore) GetAll(ctx context.Context) ([]*InstanceGroup, error) {
	g := goon.FromContext(ctx)
	r := []*InstanceGroup{}
	k := g.Kind(new(InstanceGroup))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll InstanceGroup because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *InstanceGroupStore) Get(ctx context.Context, id string) (*InstanceGroup, error) {
	g := goon.FromContext(ctx)
	r := InstanceGroup{Id: id}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get InstanceGroup because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for InstanceGroup because of %v\n", err)
		return nil, err
	}

	return &r, nil
}

func (s *InstanceGroupStore) Put(ctx context.Context, m *InstanceGroup) (*datastore.Key, error) {
	g := goon.FromContext(ctx)
	if m.Id == "" {
		m.Id = uuid.NewV4().String()
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for InstanceGroup because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *InstanceGroupStore) ValidateParent(m *InstanceGroup) error {
	if s.ParentKey == nil {
		return nil
	}
	if m.Parent == nil {
		return fmt.Errorf("No Parent given to %v", m)
	}
	if !s.ParentKey.Equal(m.Parent) {
		return fmt.Errorf("Invalid Parent for %v", m)
	}
	return nil
}
