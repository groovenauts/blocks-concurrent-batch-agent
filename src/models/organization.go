package models

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"gopkg.in/go-playground/validator.v9"
)

type (
	Organization struct {
		ID        string `datastore:"-"`
		Name			string `validate:"required"`
		Memo      string
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func (m *Organization) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Organization) Create(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Organizations", nil)
	t := time.Now()
	m.CreatedAt = t
	m.UpdatedAt = t
	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = res.Encode()
	return nil
}

func (m *Organization) Destroy(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Organization) Update(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	t := time.Now()
	m.UpdatedAt = t
	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	return nil
}
