package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type JobStatus string

const (
	Inactive        JobStatus = "inactive"
	Blocked         JobStatus = "blocked"
	Publishing      JobStatus = "publishing"
	PublishingError JobStatus = "publishing_error"
	Published       JobStatus = "published"
	Started         JobStatus = "started"
	Success         JobStatus = "success"
	Failure         JobStatus = "failure"
)

type JobKeyValuePair struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value,omitempty" datastore:"noindex"`
}

type JobKeyValuePairs []JobKeyValuePair

type JobMessage struct {
	AttributeEntries JobKeyValuePairs `json:"attribute_entries,omitempty"`
	Data             string           `json:"data,omitempty" datastore:"noindex"`
}

type Job struct {
	Id           int64          `datastore:"-" goon:"id" json:"id"`
	ParentKey    *datastore.Key `datastore:"-" goon:"parent" json:"-"`
	IDByClient   string         `json:"id_by_client" validate:"required"`
	Status       JobStatus      `json:"status" validate:"required"`
	Zone         string         `json:"zone,omitempty"`
	Hostname     string         `json:"hostname,omitempty"`
	Message      JobMessage     `json:"message,omitempty"`
	MessageID    string         `json:"message_id,omitempty"`
	Output       string         `json:"output,omitempty"`
	PipelineName string         `json:"pipeline_name,omitempty"`
	PublishedAt  time.Time      `json:"published_at,omitempty"`
	StartedAt    time.Time      `json:"started_at,omitempty"`
	FinishedAt   time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at" validate:"required"`
	UpdatedAt    time.Time      `json:"updated_at" validate:"required"`
}

func (m *Job) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *Job) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Job) Parent(ctx context.Context) (*PipelineBase, error) {
	parentStore := &PipelineBaseStore{}
	return parentStore.ByKey(ctx, m.ParentKey)
}

type JobStore struct {
	ParentKey *datastore.Key
}

func (s *JobStore) All(ctx context.Context) ([]*Job, error) {
	g := GoonFromContext(ctx)
	r := []*Job{}
	k := g.Kind(new(Job))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll Job because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *JobStore) ByID(ctx context.Context, id int64) (*Job, error) {
	r := Job{ParentKey: s.ParentKey, Id: id}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *JobStore) ByKey(ctx context.Context, key *datastore.Key) (*Job, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "JobStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := Job{ParentKey: key.Parent(), Id: key.IntID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *JobStore) Get(ctx context.Context, m *Job) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Job because of %v\n", err)
		return err
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Job because of %v\n", err)
		return err
	}

	return nil
}

func (s *JobStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&Job{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	if key.Parent() == nil {
		return fmt.Errorf("key parent must not be nil but was nil")
	}
	return nil
}

func (s *JobStore) Create(ctx context.Context, m *Job) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *JobStore) Update(ctx context.Context, m *Job) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *JobStore) ValidateAndPut(ctx context.Context, m *Job) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
}

func (s *JobStore) Put(ctx context.Context, m *Job) (*datastore.Key, error) {
	g := GoonFromContext(ctx)
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Job because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *JobStore) ValidateParent(m *Job) error {
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

func (s *JobStore) Delete(ctx context.Context, m *Job) error {
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
