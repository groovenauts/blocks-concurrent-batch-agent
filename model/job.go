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
	Id             string         `datastore:"-" goon:"id" json:"id"`
	Parent         *datastore.Key `datastore:"-" goon:"parent" json:"-"`
	IDByClient     string         `json:"id_by_client" validate:"required"`
	Status         JobStatus      `json:"status" validate:"required"`
	Zone           string         `json:"zone,omitempty"`
	Hostname       string         `json:"hostname,omitempty"`
	Message        JobMessage     `json:"message,omitempty"`
	MessageID      string         `json:"message_id,omitempty"`
	Output         string         `json:"output,omitempty"`
	PipelineId     string         `json:"pipeline_id,omitempty"`
	PipelineBaseId string         `json:"pipeline_base_id" validate:"required"`
	PublishedAt    time.Time      `json:"published_at,omitempty"`
	StartedAt      time.Time      `json:"started_at,omitempty"`
	FinishedAt     time.Time      `json:"finished_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at" validate:"required"`
	UpdatedAt      time.Time      `json:"updated_at" validate:"required"`
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

type JobStore struct {
	ParentKey *datastore.Key
}

func (s *JobStore) GetAll(ctx context.Context) ([]*Job, error) {
	g := goon.FromContext(ctx)
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

func (s *JobStore) Get(ctx context.Context, id string) (*Job, error) {
	g := goon.FromContext(ctx)
	r := Job{Id: id}
	if s.ParentKey != nil {
		r.Parent = s.ParentKey
	}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Job because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for Job because of %v\n", err)
		return nil, err
	}

	return &r, nil
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
	g := goon.FromContext(ctx)
	if m.Id == "" {
		m.Id = uuid.NewV4().String()
	}
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
	if m.Parent == nil {
		m.Parent = s.ParentKey
	}
	if !s.ParentKey.Equal(m.Parent) {
		return fmt.Errorf("Invalid Parent for %v", m)
	}
	return nil
}

func (s *JobStore) Delete(ctx context.Context, m *Job) error {
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
