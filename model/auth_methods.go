package model

import (
	"encoding/base64"
	"math/rand"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	// "google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

func (m *Auth) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Auth) Create(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}
	m.Password = m.generatePassword()

	enc_pw, err := bcrypt.GenerateFromPassword([]byte(m.Password), 10)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v\n", err)
		return err
	}
	m.EncryptedPassword = string(enc_pw) // EncryptedPassword is binary string

	store := &AuthStore{ParentKey: m.ParentKey}
	key, err := store.Create(ctx, m)
	if err != nil {
		log.Errorf(ctx, "Failed to Create %v because of %v\n", m, err)
		return err
	}
	// log.Debugf(ctx, "CreateAuth res: %v\n", res)
	keyEnc := key.Encode()
	m.Token = keyEnc + ":" + m.Password
	// log.Debugf(ctx, "CreateAuth result: %v\n", m)
	return nil
}

func (m *Auth) generatePassword() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
