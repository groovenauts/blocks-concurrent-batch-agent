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

type CloudAsyncOperationError struct {
	Code     string `json:"code" validate:"required"`
	Location string `json:"location,omitempty" `
	Message  string `json:"message,omitempty" `
}

type CloudAsyncOperationLog struct {
	Message   string    `json:"message" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

type CloudAsyncOperation struct {
	Id            string                     `datastore:"-" goon:"id" json:"id"`
	Parent        *datastore.Key             `datastore:"-" goon:"parent" json:"-"`
	OwnerType     string                     `json:"owner_type" validate:"required"`
	OwnerID       string                     `json:"owner_id" validate:"required"`
	Name          string                     `json:"name" validate:"required"`
	Service       string                     `json:"service" validate:"required"`
	OperationType string                     `json:"operation_type" validate:"required"`
	Status        string                     `json:"status" validate:"required"`
	ProjectId     string                     `json:"project_id" validate:"required"`
	Zone          string                     `json:"zone" validate:"required"`
	Errors        []CloudAsyncOperationError `json:"errors,omitempty" `
	Logs          []CloudAsyncOperationLog   `json:"logs,omitempty" `
	CreatedAt     time.Time                  `json:"created_at" validate:"required"`
	UpdatedAt     time.Time                  `json:"updated_at" validate:"required"`
}

func (m *CloudAsyncOperation) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *CloudAsyncOperation) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
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

func (s *CloudAsyncOperationStore) Create(ctx context.Context, m *CloudAsyncOperation) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *CloudAsyncOperationStore) Update(ctx context.Context, m *CloudAsyncOperation) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *CloudAsyncOperationStore) ValidateAndPut(ctx context.Context, m *CloudAsyncOperation) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
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

func (s *CloudAsyncOperationStore) Delete(ctx context.Context, m *CloudAsyncOperation) error {
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
