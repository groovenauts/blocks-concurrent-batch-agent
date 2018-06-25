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

type CloudAsyncOperation struct {
	Id            string         `datastore:"-" goon:"id"`
	Parent        *datastore.Key `datastore:"-" goon:"parent"`
	OwnerType     string
	OwnerID       string
	Name          string
	Service       string
	OperationType string
	Status        string
	ProjectId     string
	Zone          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CloudAsyncOperationStore struct {
	ParentKey *datastore.Key
}

func (s *CloudAsyncOperationStore) GetAll(ctx context.Context) ([]*CloudAsyncOperation, error) {
	g := goon.FromContext(ctx)
	r := []*CloudAsyncOperation{}
	k := g.Kind(new(CloudAsyncOperation))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll CloudAsyncOperation because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *CloudAsyncOperationStore) Get(ctx context.Context, id string) (*CloudAsyncOperation, error) {
	g := goon.FromContext(ctx)
	r := CloudAsyncOperation{Id: id}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get CloudAsyncOperation because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for CloudAsyncOperation because of %v\n", err)
		return nil, err
	}

	return &r, nil
}

func (s *CloudAsyncOperationStore) Put(ctx context.Context, m *CloudAsyncOperation) (*datastore.Key, error) {
	g := goon.FromContext(ctx)
	if m.Id == "" {
		m.Id = uuid.NewV4().String()
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for CloudAsyncOperation because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *CloudAsyncOperationStore) ValidateParent(m *CloudAsyncOperation) error {
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
