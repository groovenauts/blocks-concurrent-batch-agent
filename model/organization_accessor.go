package model

import (
	"errors"
	"fmt"
	// "runtime/debug"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type InvalidOrganizationID struct {
	ID string
}

func (e *InvalidOrganizationID) Error() string {
	return fmt.Sprintf("Invalid Organization ID: %q", e.ID)
}

type OrganizationAccessor struct {
}

var GlobalOrganizationAccessor = &OrganizationAccessor{}

var ErrNoSuchOrganization = errors.New("No such data in Organizations")

func (aa *OrganizationAccessor) Find(ctx context.Context, id string) (*Organization, error) {
	if id == "" {
		// debug.PrintStack()
		err := &InvalidOrganizationID{id}
		log.Errorf(ctx, "OrganizationAccessor#Find %v\n", err)
		return nil, err
	}

	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "OrganizationAccessor#Find %v id: %q\n", err, id)
		return nil, err
	}
	return aa.FindByKey(ctx, key)
}

func (aa *OrganizationAccessor) FindByKey(ctx context.Context, key *datastore.Key) (*Organization, error) {
	store := &OrganizationStore{}
	m, err := store.ByKey(ctx, key)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchOrganization
	case err != nil:
		log.Errorf(ctx, "OrganizationAccessor#Find %v id: %q\n", err, key.Encode())
		return nil, err
	}
	return m, nil
}
