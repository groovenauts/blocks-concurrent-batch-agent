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

type PipelineBaseStatus string

const (
	Opening               PipelineBaseStatus = "opening"
	OpeningError          PipelineBaseStatus = "opening_error"
	Hibernating           PipelineBaseStatus = "hibernating"
	Waking                PipelineBaseStatus = "waking"
	WakingError           PipelineBaseStatus = "waking_error"
	Awake                 PipelineBaseStatus = "awake"
	HibernationChecking   PipelineBaseStatus = "hibernation_checking"
	HibernationGoing      PipelineBaseStatus = "hibernation_going"
	HibernationGoingError PipelineBaseStatus = "hibernation_going_error"
	Closing               PipelineBaseStatus = "closing"
	ClosingError          PipelineBaseStatus = "closing_error"
	Closed                PipelineBaseStatus = "closed"
)

type PipelineContainer struct {
	Name             string
	Size             int
	Command          string
	Options          string
	StackdriverAgent bool
}

type PipelineBase struct {
	Id               string         `datastore:"-" goon:"id"`
	Parent           *datastore.Key `datastore:"-" goon:"parent"`
	Name             string
	InstanceGroup    InstanceGroupBody
	Container        PipelineContainer
	HibernationDelay int
	Status           PipelineBaseStatus
	IntanceGroupID   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PipelineBaseStore struct {
	ParentKey *datastore.Key
}

func (s *PipelineBaseStore) GetAll(ctx context.Context) ([]*PipelineBase, error) {
	g := goon.FromContext(ctx)
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

func (s *PipelineBaseStore) Get(ctx context.Context, id string) (*PipelineBase, error) {
	g := goon.FromContext(ctx)
	r := PipelineBase{Id: id}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get PipelineBase because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for PipelineBase because of %v\n", err)
		return nil, err
	}

	return &r, nil
}

func (s *PipelineBaseStore) Put(ctx context.Context, m *PipelineBase) (*datastore.Key, error) {
	g := goon.FromContext(ctx)
	if m.Id == "" {
		m.Id = uuid.NewV4().String()
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for PipelineBase because of %v\n", err)
		return nil, err
	}
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
	if m.Parent == nil {
		return fmt.Errorf("No Parent given to %v", m)
	}
	if !s.ParentKey.Equal(m.Parent) {
		return fmt.Errorf("Invalid Parent for %v", m)
	}
	return nil
}
