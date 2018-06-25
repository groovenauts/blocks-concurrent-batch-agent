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
	Id               string         `datastore:"-" goon:"id"`
	Parent           *datastore.Key `datastore:"-" goon:"parent"`
	Name             string
	InstanceGroup    InstanceGroupBody
	Container        PipelineContainer
	HibernationDelay int
	Status           PipelineStatus
	IntanceGroupID   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PipelineStore struct {
	ParentKey *datastore.Key
}

func (s *PipelineStore) GetAll(ctx context.Context) ([]*Pipeline, error) {
	g := goon.FromContext(ctx)
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

func (s *PipelineStore) Get(ctx context.Context, id string) (*Pipeline, error) {
	g := goon.FromContext(ctx)
	r := Pipeline{Id: id}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Pipeline because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for Pipeline because of %v\n", err)
		return nil, err
	}

	return &r, nil
}

func (s *PipelineStore) Put(ctx context.Context, m *Pipeline) (*datastore.Key, error) {
	g := goon.FromContext(ctx)
	if m.Id == "" {
		m.Id = uuid.NewV4().String()
	}
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
	if m.Parent == nil {
		return fmt.Errorf("No Parent given to %v", m)
	}
	if !s.ParentKey.Equal(m.Parent) {
		return fmt.Errorf("Invalid Parent for %v", m)
	}
	return nil
}
