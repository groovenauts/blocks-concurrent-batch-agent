package pipeline

import (
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"math/rand"
	"strings"
	"time"
)

type (
	AuthProps struct {
		Password          string `datastore:"-"`
		EncryptedPassword string
		Disabled          bool
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}

	Auth struct {
		ID    string
		Token string
		Props AuthProps
	}
)

var ErrNoSuchAuth = errors.New("No such data in Auths")

func CreateAuth(ctx context.Context) (*Auth, error) {
	mp := AuthProps{}
	mp.generatePassword()
	key := datastore.NewIncompleteKey(ctx, "Auths", nil)
	// Password is a string encoded by base64
	enc_pw, err := bcrypt.GenerateFromPassword([]byte(mp.Password), 10)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v\n", err)
		return nil, err
	}
	mp.EncryptedPassword = string(enc_pw) // EncryptedPassword is binary string
	t := time.Now()
	mp.CreatedAt = t
	mp.UpdatedAt = t
	res, err := datastore.Put(ctx, key, &mp)
	if err != nil {
		log.Errorf(ctx, "@CreateAuth %v mp: %v\n", err, &mp)
		return nil, err
	}
	id := res.Encode()
	return &Auth{ID: id, Token: id + ":" + mp.Password, Props: mp}, nil
}

func FindAuth(ctx context.Context, id string) (*Auth, error) {
	log.Debugf(ctx, "@FindAuth id: %v\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "@FindAuth %v id: %v\n", err, id)
		return nil, err
	}
	log.Debugf(ctx, "@FindAuth key: %v\n", key)
	ctx = context.WithValue(ctx, "Auth.key", key)
	m := &Auth{ID: id}
	err = datastore.Get(ctx, key, &m.Props)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchAuth
	case err != nil:
		log.Errorf(ctx, "@FindAuth %v id: %v\n", err, id)
		return nil, err
	}
	return m, nil
}

func FindAuthWithToken(ctx context.Context, token string) (*Auth, error) {
	parts := strings.SplitN(token, ":", 2)
	id := parts[0]
	pw := parts[1]
	auth, err := FindAuth(ctx, id)
	if err != nil {
		log.Errorf(ctx, "@FindAuthWithToken Auth not found %v id: %v\n", err, id)
		return nil, err
	}
	if auth.Props.Disabled {
		log.Errorf(ctx, "@FindAuthWithToken Auth is disabled. id: %v\n", id)
		return nil, err
	}
	enc_pw := auth.Props.EncryptedPassword // EncryptedPassword is binary string
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
		key, err := iter.Next(&m.Props)
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

func (m *Auth) destroy(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Auth) update(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	t := time.Now()
	m.Props.UpdatedAt = t
	_, err = datastore.Put(ctx, key, &m.Props)
	if err != nil {
		return err
	}
	return nil
}

func (mp *AuthProps) generatePassword() {
	b := make([]byte, 12)
	rand.Read(b)
	mp.Password = base64.StdEncoding.EncodeToString(b)
}
