package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type PipelineBaseStatus string

const (
	OpeningStarting       PipelineBaseStatus = "opening_starting"
	OpeningRunning        PipelineBaseStatus = "opening_running"
	OpeningError          PipelineBaseStatus = "opening_error"
	Hibernating           PipelineBaseStatus = "hibernating"
	Waking                PipelineBaseStatus = "waking"
	WakingError           PipelineBaseStatus = "waking_error"
	Awake                 PipelineBaseStatus = "awake"
	HibernationChecking   PipelineBaseStatus = "hibernation_checking"
	HibernationGoing      PipelineBaseStatus = "hibernation_going"
	HibernationGoingError PipelineBaseStatus = "hibernation_going_error"
	ClosingStarting       PipelineBaseStatus = "closing_starting"
	ClosingRunning        PipelineBaseStatus = "closing_running"
	ClosingError          PipelineBaseStatus = "closing_error"
	Closed                PipelineBaseStatus = "closed"
)

type PipelineContainer struct {
	Name             string `json:"name" validate:"required"`
	Size             int    `json:"size" validate:"required"`
	Command          string `json:"command,omitempty"`
	Options          string `json:"options,omitempty"`
	StackdriverAgent bool   `json:"stackdriver_agent,omitempty"`
}

type PipelineBase struct {
	Name             string             `datastore:"-" goon:"id" json:"name"`
	ParentKey        *datastore.Key     `datastore:"-" goon:"parent" json:"-"`
	ProjectID        string             `json:"project_id" validate:"required"`
	Zone             string             `json:"zone" validate:"required"`
	InstanceGroup    InstanceGroupBody  `json:"instance_group" validate:"required"`
	Container        PipelineContainer  `json:"container" validate:"required"`
	HibernationDelay int                `json:"hibernation_delay" validate:"required"`
	Status           PipelineBaseStatus `json:"status" validate:"required"`
	IntanceGroupName string             `json:"intance_group_name,omitempty"`
	CreatedAt        time.Time          `json:"created_at" validate:"required"`
	UpdatedAt        time.Time          `json:"updated_at" validate:"required"`
}

func (m *PipelineBase) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *PipelineBase) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *PipelineBase) Parent(ctx context.Context) (*Organization, error) {
	parentStore := &OrganizationStore{}
	return parentStore.ByKey(ctx, m.ParentKey)
}

type PipelineBaseStore struct {
	ParentKey *datastore.Key
}

func (s *PipelineBaseStore) All(ctx context.Context) ([]*PipelineBase, error) {
	g := GoonFromContext(ctx)
	r := []*PipelineBase{}
	k := g.Kind(new(PipelineBase))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll PipelineBase because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *PipelineBaseStore) ByID(ctx context.Context, name string) (*PipelineBase, error) {
	r := PipelineBase{ParentKey: s.ParentKey, Name: name}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PipelineBaseStore) ByKey(ctx context.Context, key *datastore.Key) (*PipelineBase, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "PipelineBaseStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := PipelineBase{ParentKey: key.Parent(), Name: key.StringID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *PipelineBaseStore) Get(ctx context.Context, m *PipelineBase) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get PipelineBase because of %v\n", err)
		return err
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for PipelineBase because of %v\n", err)
		return err
	}

	return nil
}

func (s *PipelineBaseStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&PipelineBase{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	if key.Parent() == nil {
		return fmt.Errorf("key parent must not be nil but was nil")
	}
	return nil
}

func (s *PipelineBaseStore) Exist(ctx context.Context, m *PipelineBase) (bool, error) {
	g := GoonFromContext(ctx)
	key, err := g.KeyError(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Key of %v because of %v\n", m, err)
		return false, err
	}
	_, err = s.ByKey(ctx, key)
	if err == datastore.ErrNoSuchEntity {
		return false, nil
	} else if err != nil {
		log.Errorf(ctx, "Failed to get existance of %v because of %v\n", m, err)
		return false, err
	} else {
		return true, nil
	}
}

func (s *PipelineBaseStore) Create(ctx context.Context, m *PipelineBase) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	exist, err := s.Exist(ctx, m)
	if err != nil {
		return nil, err
	}
	if exist {
		log.Errorf(ctx, "Failed to create %v because of another entity has same key\n", m)
		return nil, fmt.Errorf("Duplicate Name error: %q of %v\n", m.Name, m)
	}

	return s.Put(ctx, m)
}

func (s *PipelineBaseStore) Update(ctx context.Context, m *PipelineBase) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	exist, err := s.Exist(ctx, m)
	if err != nil {
		return nil, err
	}
	if !exist {
		log.Errorf(ctx, "Failed to update %v because it doesn't exist\n", m)
		return nil, fmt.Errorf("No data to update %q of %v\n", m.Name, m)
	}

	return s.Put(ctx, m)
}

func (s *PipelineBaseStore) Put(ctx context.Context, m *PipelineBase) (*datastore.Key, error) {
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for PipelineBase because of %v\n", err)
		return nil, err
	}
	g := GoonFromContext(ctx)
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *PipelineBaseStore) ValidateParent(m *PipelineBase) error {
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

func (s *PipelineBaseStore) Delete(ctx context.Context, m *PipelineBase) error {
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
