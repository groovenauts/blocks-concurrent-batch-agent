package models

import (
	"encoding/base64"
	"math/rand"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type (
	Auth struct {
		ID                string `datastore:"-"`
		Token             string `datastore:"-"`
		Password          string `datastore:"-"`
		EncryptedPassword string
		Disabled          bool
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}
)

func (m *Auth) Destroy(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Auth) Update(ctx context.Context) error {
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

func (m *Auth) generatePassword() {
	b := make([]byte, 12)
	rand.Read(b)
	m.Password = base64.StdEncoding.EncodeToString(b)
}
