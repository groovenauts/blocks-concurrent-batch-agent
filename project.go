package pipeline

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

type Project struct {
	ID        string    `json:"id"   datastore:"-"`
	Name      string    `json:"name" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *Project) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Project) Key() (*datastore.Key, error) {
	return datastore.DecodeKey(m.ID)
}

func (m *Project) update(ctx context.Context) error {
	t := time.Now()
	m.UpdatedAt = t

	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := m.Key()
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Project) destroy(ctx context.Context) error {
	key, err := m.Key()
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}
