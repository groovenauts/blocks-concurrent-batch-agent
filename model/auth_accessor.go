package model

import (
	"fmt"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type InvalidParent struct {
	ID string
}

func (e *InvalidParent) Error() string {
	return fmt.Sprintf("Invalid parent from ID: %q", e.ID)
}


type AuthAccessor struct {
	Parent *Organization
}

var GlobalAuthAccessor = &AuthAccessor{}

var ErrNoSuchAuth = errors.New("No such data in Auths")


func (aa *AuthAccessor) Find(ctx context.Context, id string) (*Auth, error) {
	// log.Debugf(ctx, "@FindAuth id: %q\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "@FindAuth %v id: %q\n", err, id)
		return nil, err
	}
	if aa.Parent != nil {
		parentKey, err := datastore.DecodeKey(aa.Parent.ID)
		if err != nil {
			return nil, err
		}
		if !parentKey.Equal(key.Parent()) {
			return nil, &InvalidParent{id}
		}
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

func (aa *AuthAccessor) FindWithToken(ctx context.Context, token string) (*Auth, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		err := errors.New("Invalid token: " + token)
		log.Errorf(ctx, "@FindAuthWithToken %v", err)
		return nil, err
	}
	id := parts[0]
	pw := parts[1]
	auth, err := aa.Find(ctx, id)
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

func (aa *AuthAccessor) GetAll(ctx context.Context) ([]*Auth, error) {
	q := datastore.NewQuery("Auths")
	if aa.Parent != nil {
		key, err := datastore.DecodeKey(aa.Parent.ID)
		if err != nil {
			return nil, err
		}
		q = q.Ancestor(key)
	}
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
