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
	ParentKey         *datastore.Key `datastore:"-" goon:"parent" json:"-"`
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

func (m *Auth) Parent(ctx context.Context) (*Organization, error) {
	parentStore := &OrganizationStore{}
	return parentStore.ByKey(ctx, m.ParentKey)
}

type AuthStore struct {
	ParentKey *datastore.Key
}

func (s *AuthStore) All(ctx context.Context) ([]*Auth, error) {
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

func (s *AuthStore) ByID(ctx context.Context, iD int64) (*Auth, error) {
	r := Auth{ParentKey: s.ParentKey, ID: iD}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *AuthStore) ByKey(ctx context.Context, key *datastore.Key) (*Auth, error) {
	if err := s.IsValidKey(ctx, key); err != nil {
		log.Errorf(ctx, "AuthStore.ByKey got Invalid key: %v because of %v\n", key, err)
		return nil, err
	}

	r := Auth{ParentKey: key.Parent(), ID: key.IntID()}
	err := s.Get(ctx, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *AuthStore) Get(ctx context.Context, m *Auth) error {
	g := GoonFromContext(ctx)
	err := g.Get(m)
	if err != nil {
		log.Errorf(ctx, "Failed to Get Auth because of %v\n", err)
		return err
	}
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Auth because of %v\n", err)
		return err
	}

	return nil
}

func (s *AuthStore) IsValidKey(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}
	g := GoonFromContext(ctx)
	expected := g.Kind(&Auth{})
	if key.Kind() != expected {
		return fmt.Errorf("key kind must be %s but was %s", expected, key.Kind())
	}
	if key.Parent() == nil {
		return fmt.Errorf("key parent must not be nil but was nil")
	}
	return nil
}

func (s *AuthStore) Exist(ctx context.Context, m *Auth) (bool, error) {
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

func (s *AuthStore) Create(ctx context.Context, m *Auth) (*datastore.Key, error) {
	err := m.PrepareToCreate()
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	return s.Put(ctx, m)
}

func (s *AuthStore) Update(ctx context.Context, m *Auth) (*datastore.Key, error) {
	err := m.PrepareToUpdate()
	if err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	return s.Put(ctx, m)
}

func (s *AuthStore) Put(ctx context.Context, m *Auth) (*datastore.Key, error) {
	if err := s.ValidateParent(m); err != nil {
		log.Errorf(ctx, "Invalid parent key for Auth because of %v\n", err)
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

func (s *AuthStore) ValidateParent(m *Auth) error {
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
