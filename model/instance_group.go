package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/mjibson/goon"
)

type InstanceGroupStatus string

const (
	ConstructionStarting InstanceGroupStatus = "construction_starting"
	ConstructionRunning  InstanceGroupStatus = "construction_running"
	ConstructionError    InstanceGroupStatus = "construction_error"
	Constructed          InstanceGroupStatus = "constructed"
	HealthCheckError     InstanceGroupStatus = "health_check_error"
	ResizeStarting       InstanceGroupStatus = "resize_starting"
	ResizeRunning        InstanceGroupStatus = "resize_running"
	ResizeWaiting        InstanceGroupStatus = "resize_waiting"
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

type InstanceGroupHealthCheckConfig struct {
	Interval                 int `json:"interval,omitempty"`
	MinimumRunningSize       int `json:"minimum_running_size,omitempty"`
	MinimumRunningPercentage int `json:"minimum_running_percentage,omitempty"`
}

type InstanceGroupBody struct {
	BootDisk              InstanceGroupVMDisk            `json:"boot_disk" validate:"required"`
	MachineType           string                         `json:"machine_type" validate:"required"`
	GpuAccelerators       InstanceGroupAccelerators      `json:"gpu_accelerators,omitempty"`
	HealthCheck           InstanceGroupHealthCheckConfig `json:"health_check,omitempty"`
	Preemptible           bool                           `json:"preemptible,omitempty"`
	InstanceSizeRequested int                            `json:"instance_size_requested,omitempty"`
	StartupScript         string                         `json:"startup_script,omitempty"`
	DeploymentName        string                         `json:"deployment_name,omitempty"`
	TokenConsumption      int                            `json:"token_consumption,omitempty"`
}

type InstanceGroup struct {
	Id                    int64                          `datastore:"-" goon:"id" json:"id"`
	Parent                *datastore.Key                 `datastore:"-" goon:"parent" json:"-"`
	Name                  string                         `json:"name" validate:"required"`
	ProjectID             string                         `json:"project_id" validate:"required"`
	Zone                  string                         `json:"zone" validate:"required"`
	BootDisk              InstanceGroupVMDisk            `json:"boot_disk" validate:"required"`
	MachineType           string                         `json:"machine_type" validate:"required"`
	GpuAccelerators       InstanceGroupAccelerators      `json:"gpu_accelerators,omitempty"`
	HealthCheck           InstanceGroupHealthCheckConfig `json:"health_check,omitempty"`
	Preemptible           bool                           `json:"preemptible,omitempty"`
	InstanceSizeRequested int                            `json:"instance_size_requested,omitempty"`
	StartupScript         string                         `json:"startup_script,omitempty"`
	DeploymentName        string                         `json:"deployment_name,omitempty"`
	TokenConsumption      int                            `json:"token_consumption,omitempty"`
	InstanceSize          int                            `json:"instance_size,omitempty"`
	HealthCheckId         string                         `json:"health_check_id,omitempty"`
	Status                InstanceGroupStatus            `json:"status" validate:"required"`
	CreatedAt             time.Time                      `json:"created_at" validate:"required"`
	UpdatedAt             time.Time                      `json:"updated_at" validate:"required"`
}

type InstanceGroupHealthCheck struct {
	Id         int64          `datastore:"-" goon:"id" json:"id"`
	Parent     *datastore.Key `datastore:"-" goon:"parent" json:"-"`
	LastResult string         `json:"last_result,omitempty"`
	CreatedAt  time.Time      `json:"created_at" validate:"required"`
	UpdatedAt  time.Time      `json:"updated_at" validate:"required"`
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

func (m *InstanceGroupHealthCheck) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *InstanceGroupHealthCheck) PrepareToUpdate() error {
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

func (s *InstanceGroupStore) Get(ctx context.Context, id int64) (*InstanceGroup, error) {
	g := goon.FromContext(ctx)
	r := InstanceGroup{Id: id}
	if s.ParentKey != nil {
		r.Parent = s.ParentKey
	}
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
		m.Parent = s.ParentKey
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

type InstanceGroupHealthCheckStore struct {
	ParentKey *datastore.Key
}

func (s *InstanceGroupHealthCheckStore) GetAll(ctx context.Context) ([]*InstanceGroupHealthCheck, error) {
	g := goon.FromContext(ctx)
	r := []*InstanceGroupHealthCheck{}
	k := g.Kind(new(InstanceGroupHealthCheck))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll InstanceGroupHealthCheck because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *InstanceGroupHealthCheckStore) Get(ctx context.Context, id int64) (*InstanceGroupHealthCheck, error) {
	g := goon.FromContext(ctx)
	r := InstanceGroupHealthCheck{Id: id}
	if s.ParentKey != nil {
		r.Parent = s.ParentKey
	}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get InstanceGroupHealthCheck because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for InstanceGroupHealthCheck because of %v\n", err)
		return nil, err
	}

	return &r, nil
}

func (s *InstanceGroupHealthCheckStore) Create(ctx context.Context, m *InstanceGroupHealthCheck) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *InstanceGroupHealthCheckStore) Update(ctx context.Context, m *InstanceGroupHealthCheck) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *InstanceGroupHealthCheckStore) ValidateAndPut(ctx context.Context, m *InstanceGroupHealthCheck) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
}

func (s *InstanceGroupHealthCheckStore) Put(ctx context.Context, m *InstanceGroupHealthCheck) (*datastore.Key, error) {
	g := goon.FromContext(ctx)
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for InstanceGroupHealthCheck because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *InstanceGroupHealthCheckStore) ValidateParent(m *InstanceGroupHealthCheck) error {
	if s.ParentKey == nil {
		return nil
	}
	if m.Parent == nil {
		m.Parent = s.ParentKey
	}
	if !s.ParentKey.Equal(m.Parent) {
		return fmt.Errorf("Invalid Parent for %v", m)
	}
	return nil
}

func (s *InstanceGroupHealthCheckStore) Delete(ctx context.Context, m *InstanceGroupHealthCheck) error {
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
