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
	DiskSizeGb  int    `json:"disk_size_gb,omitempty"`
	DiskType    string `json:"disk_type,omitempty"`
	SourceImage string `json:"source_image" validate:"required"`
}

type InstanceGroupAccelerators struct {
	Count int    `json:"count,omitempty"`
	Type  string `json:"type,omitempty"`
}

type InstanceGroupBody struct {
	ProjectID             string                    `json:"project_id" validate:"required"`
	Zone                  string                    `json:"zone" validate:"required"`
	BootDisk              InstanceGroupVMDisk       `json:"boot_disk" validate:"required"`
	MachineType           string                    `json:"machine_type" validate:"required"`
	GpuAccelerators       InstanceGroupAccelerators `json:"gpu_accelerators,omitempty"`
	Preemptible           bool                      `json:"preemptible,omitempty"`
	InstanceSizeRequested int                       `json:"instance_size_requested,omitempty"`
	InstanceSize          int                       `json:"instance_size,omitempty"`
	StartupScript         string                    `json:"startup_script,omitempty"`
	Status                InstanceGroupStatus       `json:"status" validate:"required"`
	DeploymentName        string                    `json:"deployment_name,omitempty"`
	TokenConsumption      int                       `json:"token_consumption,omitempty"`
}

type InstanceGroup struct {
	Id                    string                    `datastore:"-" goon:"id" json:"id"`
	Parent                *datastore.Key            `datastore:"-" goon:"parent" json:"-"`
	Name                  string                    `json:"name" validate:"required"`
	ProjectID             string                    `json:"project_id" validate:"required"`
	Zone                  string                    `json:"zone" validate:"required"`
	BootDisk              InstanceGroupVMDisk       `json:"boot_disk" validate:"required"`
	MachineType           string                    `json:"machine_type" validate:"required"`
	GpuAccelerators       InstanceGroupAccelerators `json:"gpu_accelerators,omitempty"`
	Preemptible           bool                      `json:"preemptible,omitempty"`
	InstanceSizeRequested int                       `json:"instance_size_requested,omitempty"`
	InstanceSize          int                       `json:"instance_size,omitempty"`
	StartupScript         string                    `json:"startup_script,omitempty"`
	Status                InstanceGroupStatus       `json:"status" validate:"required"`
	DeploymentName        string                    `json:"deployment_name,omitempty"`
	TokenConsumption      int                       `json:"token_consumption,omitempty"`
	CreatedAt             time.Time                 `json:"created_at" validate:"required"`
	UpdatedAt             time.Time                 `json:"updated_at" validate:"required"`
}

func (m *InstanceGroup) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *InstanceGroup) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
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

func (s *InstanceGroupStore) Create(ctx context.Context, m *InstanceGroup) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *InstanceGroupStore) Update(ctx context.Context, m *InstanceGroup) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *InstanceGroupStore) ValidateAndPut(ctx context.Context, m *InstanceGroup) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
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

func (s *InstanceGroupStore) Delete(ctx context.Context, m *InstanceGroup) error {
	g := goon.FromContext(ctx)
	key, err := g.KeyError(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get %v because of %v\n", m, err)
		return err
	}
	err = g.Delete(key)
	if err != nil {
		log.Errorf(ctx, "Failed to Delete %v because of %v\n", m, err)
		return err
	}
	return nil
}
