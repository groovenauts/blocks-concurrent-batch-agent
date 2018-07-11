package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Auth struct {
	ID                int64          `datastore:"-" goon:"id" json:"id"`
	Parent            *datastore.Key `datastore:"-" goon:"parent" json:"-"`
	Token             string         `json:"token,omitempty" datastore:"-"`
	Password          string         `json:"password,omitempty" datastore:"-"`
	EncryptedPassword string         `json:"encrypted_password,omitempty"`
	Disabled          bool           `json:"disabled,omitempty"`
	CreatedAt         time.Time      `json:"created_at" validate:"required"`
	UpdatedAt         time.Time      `json:"updated_at" validate:"required"`
}

func (m *Auth) PrepareToCreate() error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}
	return nil
}

func (m *Auth) PrepareToUpdate() error {
	m.UpdatedAt = time.Now()
	return nil
}

type AuthStore struct {
	ParentKey *datastore.Key
}

func (s *AuthStore) GetAll(ctx context.Context) ([]*Auth, error) {
	g := GoonFromContext(ctx)
	r := []*Auth{}
	k := g.Kind(new(Auth))
	log.Infof(ctx, "Kind is %v\n", k)
	q := datastore.NewQuery(k)
	q = q.Ancestor(s.ParentKey)
	log.Infof(ctx, "q is %v\n", q)
	_, err := g.GetAll(q.EventualConsistency(), &r)
	if err != nil {
		log.Errorf(ctx, "Failed to GetAll Auth because of %v\n", err)
		return nil, err
	}
	return r, nil
}

func (s *AuthStore) Get(ctx context.Context, iD int64) (*Auth, error) {
	g := GoonFromContext(ctx)
	r := Auth{ID: iD}
	if s.ParentKey != nil {
		r.Parent = s.ParentKey
	}
	err := g.Get(&r)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Auth because of %v\n", err)
		return nil, err
	}
	if err := s.ValidateParent(&r); err != nil {
		log.Errorf(ctx, "Invalid parent key for Auth because of %v\n", err)
		return nil, err
	}

	return &r, nil
}

func (s *AuthStore) Create(ctx context.Context, m *Auth) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *AuthStore) Update(ctx context.Context, m *Auth) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	return s.ValidateAndPut(ctx, m)
}

func (s *AuthStore) ValidateAndPut(ctx context.Context, m *Auth) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}
	return s.Put(ctx, m)
}

func (s *AuthStore) Put(ctx context.Context, m *Auth) (*datastore.Key, error) {
	g := GoonFromContext(ctx)
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Auth because of %v\n", err)
		return nil, err
	}
	key, err := g.Put(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Put %v because of %v\n", m, err)
		return nil, err
	}
	return key, nil
}

func (s *AuthStore) ValidateParent(m *Auth) error {
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

func (s *AuthStore) Delete(ctx context.Context, m *Auth) error {
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
