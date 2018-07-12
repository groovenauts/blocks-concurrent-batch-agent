package model

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

func (m *Organization) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Organization) Create(ctx context.Context) (*datastore.Key, error) {
	orgStore := &OrganizationStore{}
	return orgStore.Create(ctx, m)
}
