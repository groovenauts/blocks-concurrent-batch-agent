package models

import (
	"encoding/base64"
	"errors"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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

var ErrNoSuchAuth = errors.New("No such data in Auths")

func CreateAuth(ctx context.Context) (*Auth, error) {
	m := Auth{}
	m.generatePassword()
	key := datastore.NewIncompleteKey(ctx, "Auths", nil)
	// Password is a string encoded by base64
	enc_pw, err := bcrypt.GenerateFromPassword([]byte(m.Password), 10)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v\n", err)
		return nil, err
	}
	m.EncryptedPassword = string(enc_pw) // EncryptedPassword is binary string
	t := time.Now()
	m.CreatedAt = t
	m.UpdatedAt = t
	res, err := datastore.Put(ctx, key, &m)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v mp: %v\n", err, &m)
		return nil, err
	}
	// log.Debugf(ctx, "CreateAuth res: %v\n", res)
	id := res.Encode()
	m.ID = id
	m.Token = id + ":" + m.Password
	// log.Debugf(ctx, "CreateAuth result: %v\n", m)
	return &m, nil
}

func FindAuth(ctx context.Context, id string) (*Auth, error) {
	// log.Debugf(ctx, "@FindAuth id: %q\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "@FindAuth %v id: %q\n", err, id)
		return nil, err
	}
	// log.Debugf(ctx, "@FindAuth key: %q\n", key)
	ctx = context.WithValue(ctx, "Auth.key", key)
	m := &Auth{ID: id}
	err = datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchAuth
	case err != nil:
		log.Errorf(ctx, "@FindAuth %v id: %q\n", err, id)
		return nil, err
	}
	return m, nil
}

func FindAuthWithToken(ctx context.Context, token string) (*Auth, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		err := errors.New("Invalid token: " + token)
		log.Errorf(ctx, "@FindAuthWithToken %v", err)
		return nil, err
	}
	id := parts[0]
	pw := parts[1]
	auth, err := FindAuth(ctx, id)
	if err != nil {
		log.Errorf(ctx, "@FindAuthWithToken Auth not found %v id: %v\n", err, id)
		return nil, err
	}
	if auth.Disabled {
		log.Errorf(ctx, "@FindAuthWithToken Auth is disabled. id: %v\n", id)
		return nil, err
	}
	enc_pw := auth.EncryptedPassword // EncryptedPassword is binary string
	if err = bcrypt.CompareHashAndPassword([]byte(enc_pw), []byte(pw)); err != nil {
		log.Errorf(ctx, "@FindAuthWithToken Auth is disabled. id: %v\n", id)
		return nil, err
	}
	return auth, nil
}

func GetAllAuth(ctx context.Context) ([]*Auth, error) {
	q := datastore.NewQuery("Auths")
	iter := q.Run(ctx)
	var res = []*Auth{}
	for {
		m := Auth{}
		key, err := iter.Next(&m)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		m.ID = key.Encode()
		res = append(res, &m)
	}
	return res, nil
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