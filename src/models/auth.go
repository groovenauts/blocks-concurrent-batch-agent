package models

import (
	"context"
	"encoding/base64"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

type (
	Auth struct {
		ID                string        `datastore:"-"`
		Organization      *Organization `datastore:"-" validate:"required"`
		Token             string        `datastore:"-"`
		Password          string        `datastore:"-"`
		EncryptedPassword string
		Disabled          bool
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}
)

func (m *Auth) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Auth) Create(ctx context.Context) error {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = t
	}

	err := m.Validate()
	if err != nil {
		return err
	}
	m.generatePassword()

	orgKey, err := datastore.DecodeKey(m.Organization.ID)
	if err != nil {
		return err
	}
	key := datastore.NewIncompleteKey(ctx, "Auths", orgKey)
	// Password is a string encoded by base64
	enc_pw, err := bcrypt.GenerateFromPassword([]byte(m.Password), 10)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v\n", err)
		return err
	}
	m.EncryptedPassword = string(enc_pw) // EncryptedPassword is binary string
	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v mp: %v\n", err, m)
		return err
	}
	// log.Debugf(ctx, "CreateAuth res: %v\n", res)
	id := res.Encode()
	m.ID = id
	m.Token = id + ":" + m.Password
	// log.Debugf(ctx, "CreateAuth result: %v\n", m)
	return nil
}

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
	m.UpdatedAt = time.Now()

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
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
