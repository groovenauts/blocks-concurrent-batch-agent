package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type PipelineStatus string

const (
	CurrentPreparing      PipelineStatus = "current_preparing"
	CurrentPreparingError PipelineStatus = "current_preparing_error"
	Running               PipelineStatus = "running"
	NextPreparing         PipelineStatus = "next_preparing"
	Stopping              PipelineStatus = "stopping"
	StoppingError         PipelineStatus = "stopping_error"
	Stopped               PipelineStatus = "stopped"
)

type Pipeline struct {
	Name             string            `datastore:"-" goon:"id" json:"name"`
	ParentKey        *datastore.Key    `datastore:"-" goon:"parent" json:"-"`
	ProjectID        string            `json:"project_id" validate:"required"`
	Zone             string            `json:"zone" validate:"required"`
	InstanceGroup    InstanceGroupBody `json:"instance_group,omitempty"`
	Container        PipelineContainer `json:"container,omitempty"`
	HibernationDelay int               `json:"hibernation_delay" validate:"required"`
	Status           PipelineStatus    `json:"status" validate:"required"`
	IntanceGroupID   string            `json:"intance_group_id" validate:"required"`
	CreatedAt        time.Time         `json:"created_at" validate:"required"`
	UpdatedAt        time.Time         `json:"updated_at" validate:"required"`
}

func (m *Pipeline) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *Pipeline) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Pipeline) Parent(ctx context.Context) (*Organization, error) {
	parentStore := &OrganizationStore{}
	return parentStore.ByKey(ctx, m.ParentKey)
}

type PipelineStore struct {
	ParentKey *datastore.Key
}

func (s *PipelineStore) All(ctx context.Context) ([]*Pipeline, error) {
	g := GoonFromContext(ctx)
	r := []*Pipeline{}
	k := g.Kind(new(Pipeline))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll Pipeline because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *PipelineStore) ByID(ctx context.Context, name string) (*Pipeline, error) {
	r := Pipeline{ParentKey: s.ParentKey, Name: name}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PipelineStore) ByKey(ctx context.Context, key *datastore.Key) (*Pipeline, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "PipelineStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := Pipeline{ParentKey: key.Parent(), Name: key.StringID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PipelineStore) Get(ctx context.Context, m *Pipeline) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Pipeline because of %v\n", err)
		return err
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Pipeline because of %v\n", err)
		return err
	}

	return nil
}

func (s *PipelineStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&Pipeline{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	if key.Parent() == nil {
		return fmt.Errorf("key parent must not be nil but was nil")
	}
	return nil
}

func (s *PipelineStore) Create(ctx context.Context, m *Pipeline) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *PipelineStore) Update(ctx context.Context, m *Pipeline) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *PipelineStore) ValidateAndPut(ctx context.Context, m *Pipeline) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
}

func (s *PipelineStore) Put(ctx context.Context, m *Pipeline) (*datastore.Key, error) {
	g := GoonFromContext(ctx)
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Pipeline because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *PipelineStore) ValidateParent(m *Pipeline) error {
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

func (s *PipelineStore) Delete(ctx context.Context, m *Pipeline) error {
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
