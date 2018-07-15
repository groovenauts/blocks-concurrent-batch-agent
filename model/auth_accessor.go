package model

import (
	"errors"
	"fmt"
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

type AuthAccessor struct{}

func (aa *AuthAccessor) FindWithToken(ctx context.Context, token string) (*Auth, error) {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 {
		err := errors.New("Invalid token: " + token)
		log.Errorf(ctx, "FindAuthWithToken %v", err)
		return nil, err
	}
	keyEnc := parts[0]
	pw := parts[1]
	key, err := datastore.DecodeKey(keyEnc)
	if err != nil {
		log.Errorf(ctx, "FindAuthWithToken Invalid keyEnc: %v because of \n", keyEnc, err)
	}
	store := &AuthStore{}
	auth, err := store.ByKey(ctx, key)
	if err != nil {
		log.Errorf(ctx, "FindAuthWithToken Auth not found %v keyEnc: %v\n", err, keyEnc)
		return nil, err
	}
	if auth.Disabled {
		log.Errorf(ctx, "FindAuthWithToken Auth is disabled. keyEnc: %v\n", keyEnc)
		return nil, err
	}
	enc_pw := auth.EncryptedPassword // EncryptedPassword is binary string
	if err = bcrypt.CompareHashAndPassword([]byte(enc_pw), []byte(pw)); err != nil {
		log.Errorf(ctx, "FindAuthWithToken Auth is disabled. keyEnc: %v\n", keyEnc)
		return nil, err
	}
	return auth, nil
}
