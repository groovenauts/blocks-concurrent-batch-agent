package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type CloudAsyncOperationError struct {
	Code     string `json:"code" validate:"required"`
	Location string `json:"location,omitempty"`
	Message  string `json:"message,omitempty"`
}

type CloudAsyncOperationLog struct {
	Message   string    `json:"message" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

type InstanceGroupOperation struct {
	Id            int64                      `datastore:"-" goon:"id" json:"id"`
	ParentKey     *datastore.Key             `datastore:"-" goon:"parent" json:"-"`
	Name          string                     `json:"name" validate:"required"`
	Service       string                     `json:"service" validate:"required"`
	OperationType string                     `json:"operation_type" validate:"required"`
	Status        string                     `json:"status" validate:"required"`
	ProjectId     string                     `json:"project_id" validate:"required"`
	Zone          string                     `json:"zone" validate:"required"`
	Errors        []CloudAsyncOperationError `json:"errors,omitempty"`
	Logs          []CloudAsyncOperationLog   `json:"logs,omitempty"`
	CreatedAt     time.Time                  `json:"created_at" validate:"required"`
	UpdatedAt     time.Time                  `json:"updated_at" validate:"required"`
}

type PipelineBaseOperation struct {
	Id            int64                      `datastore:"-" goon:"id" json:"id"`
	ParentKey     *datastore.Key             `datastore:"-" goon:"parent" json:"-"`
	Name          string                     `json:"name" validate:"required"`
	Service       string                     `json:"service" validate:"required"`
	OperationType string                     `json:"operation_type" validate:"required"`
	Status        string                     `json:"status" validate:"required"`
	ProjectId     string                     `json:"project_id" validate:"required"`
	Zone          string                     `json:"zone" validate:"required"`
	Errors        []CloudAsyncOperationError `json:"errors,omitempty"`
	Logs          []CloudAsyncOperationLog   `json:"logs,omitempty"`
	CreatedAt     time.Time                  `json:"created_at" validate:"required"`
	UpdatedAt     time.Time                  `json:"updated_at" validate:"required"`
}

func (m *InstanceGroupOperation) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *InstanceGroupOperation) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *InstanceGroupOperation) Parent(ctx context.Context) (*InstanceGroup, error) {
	parentStore := &InstanceGroupStore{}
	return parentStore.ByKey(ctx, m.ParentKey)
}

func (m *PipelineBaseOperation) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *PipelineBaseOperation) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *PipelineBaseOperation) Parent(ctx context.Context) (*PipelineBase, error) {
	parentStore := &PipelineBaseStore{}
	return parentStore.ByKey(ctx, m.ParentKey)
}

type InstanceGroupOperationStore struct {
	ParentKey *datastore.Key
}

func (s *InstanceGroupOperationStore) All(ctx context.Context) ([]*InstanceGroupOperation, error) {
	g := GoonFromContext(ctx)
	r := []*InstanceGroupOperation{}
	k := g.Kind(new(InstanceGroupOperation))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll InstanceGroupOperation because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *InstanceGroupOperationStore) ByID(ctx context.Context, id int64) (*InstanceGroupOperation, error) {
	r := InstanceGroupOperation{ParentKey: s.ParentKey, Id: id}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *InstanceGroupOperationStore) ByKey(ctx context.Context, key *datastore.Key) (*InstanceGroupOperation, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "InstanceGroupOperationStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := InstanceGroupOperation{ParentKey: key.Parent(), Id: key.IntID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *InstanceGroupOperationStore) Get(ctx context.Context, m *InstanceGroupOperation) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get InstanceGroupOperation because of %v\n", err)
		return err
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for InstanceGroupOperation because of %v\n", err)
		return err
	}

	return nil
}

func (s *InstanceGroupOperationStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&InstanceGroupOperation{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	if key.Parent() == nil {
		return fmt.Errorf("key parent must not be nil but was nil")
	}
	return nil
}

func (s *InstanceGroupOperationStore) Create(ctx context.Context, m *InstanceGroupOperation) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *InstanceGroupOperationStore) Update(ctx context.Context, m *InstanceGroupOperation) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *InstanceGroupOperationStore) ValidateAndPut(ctx context.Context, m *InstanceGroupOperation) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
}

func (s *InstanceGroupOperationStore) Put(ctx context.Context, m *InstanceGroupOperation) (*datastore.Key, error) {
	g := GoonFromContext(ctx)
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for InstanceGroupOperation because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *InstanceGroupOperationStore) ValidateParent(m *InstanceGroupOperation) error {
	if s.ParentKey == nil {
		return nil
	}
	if m.ParentKey == nil {
		m.ParentKey = s.ParentKey
	}
	if !s.ParentKey.Equal(m.ParentKey) {
		return fmt.Errorf("Invalid ParentKey for %v", m)
	}
	return nil
}

func (s *InstanceGroupOperationStore) Delete(ctx context.Context, m *InstanceGroupOperation) error {
	g := GoonFromContext(ctx)
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

type PipelineBaseOperationStore struct {
	ParentKey *datastore.Key
}

func (s *PipelineBaseOperationStore) All(ctx context.Context) ([]*PipelineBaseOperation, error) {
	g := GoonFromContext(ctx)
	r := []*PipelineBaseOperation{}
	k := g.Kind(new(PipelineBaseOperation))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll PipelineBaseOperation because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *PipelineBaseOperationStore) ByID(ctx context.Context, id int64) (*PipelineBaseOperation, error) {
	r := PipelineBaseOperation{ParentKey: s.ParentKey, Id: id}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PipelineBaseOperationStore) ByKey(ctx context.Context, key *datastore.Key) (*PipelineBaseOperation, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "PipelineBaseOperationStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := PipelineBaseOperation{ParentKey: key.Parent(), Id: key.IntID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PipelineBaseOperationStore) Get(ctx context.Context, m *PipelineBaseOperation) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get PipelineBaseOperation because of %v\n", err)
		return err
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for PipelineBaseOperation because of %v\n", err)
		return err
	}

	return nil
}

func (s *PipelineBaseOperationStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&PipelineBaseOperation{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	if key.Parent() == nil {
		return fmt.Errorf("key parent must not be nil but was nil")
	}
	return nil
}

func (s *PipelineBaseOperationStore) Create(ctx context.Context, m *PipelineBaseOperation) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *PipelineBaseOperationStore) Update(ctx context.Context, m *PipelineBaseOperation) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *PipelineBaseOperationStore) ValidateAndPut(ctx context.Context, m *PipelineBaseOperation) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
}

func (s *PipelineBaseOperationStore) Put(ctx context.Context, m *PipelineBaseOperation) (*datastore.Key, error) {
	g := GoonFromContext(ctx)
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for PipelineBaseOperation because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *PipelineBaseOperationStore) ValidateParent(m *PipelineBaseOperation) error {
	if s.ParentKey == nil {
		return nil
	}
	if m.ParentKey == nil {
		m.ParentKey = s.ParentKey
	}
	if !s.ParentKey.Equal(m.ParentKey) {
		return fmt.Errorf("Invalid ParentKey for %v", m)
	}
	return nil
}

func (s *PipelineBaseOperationStore) Delete(ctx context.Context, m *PipelineBaseOperation) error {
	g := GoonFromContext(ctx)
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
