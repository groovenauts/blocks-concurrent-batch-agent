package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Organization struct {
	ID          int64     `datastore:"-" goon:"id" json:"id"`
	Name        string    `json:"name" validate:"required"`
	Memo        string    `json:"memo,omitempty"`
	TokenAmount int       `json:"token_amount,omitempty"`
	CreatedAt   time.Time `json:"created_at" validate:"required"`
	UpdatedAt   time.Time `json:"updated_at" validate:"required"`
}

func (m *Organization) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *Organization) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

type OrganizationStore struct {
}

func (s *OrganizationStore) All(ctx context.Context) ([]*Organization, error) {
	g := GoonFromContext(ctx)
	r := []*Organization{}
	k := g.Kind(new(Organization))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll Organization because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *OrganizationStore) ByID(ctx context.Context, iD int64) (*Organization, error) {
	r := Organization{ID: iD}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *OrganizationStore) ByKey(ctx context.Context, key *datastore.Key) (*Organization, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "OrganizationStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := Organization{ID: key.IntID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *OrganizationStore) Get(ctx context.Context, m *Organization) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Organization because of %v\n", err)
		return err
	}

	return nil
}

func (s *OrganizationStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&Organization{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	return nil
}

func (s *OrganizationStore) Exist(ctx context.Context, m *Organization) (bool, error) {
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

func (s *OrganizationStore) Create(ctx context.Context, m *Organization) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	return s.Put(ctx, m)
}

func (s *OrganizationStore) Update(ctx context.Context, m *Organization) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	return s.Put(ctx, m)
}

func (s *OrganizationStore) Put(ctx context.Context, m *Organization) (*datastore.Key, error) {
	g := GoonFromContext(ctx)
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}
