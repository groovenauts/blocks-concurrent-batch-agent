package models

import (
	"context"
	"errors"
	"fmt"
	// "runtime/debug"

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
	ctx = context.WithValue(ctx, "Organization.key", key)
	m := &Organization{}
	err := datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchOrganization
	case err != nil:
		log.Errorf(ctx, "OrganizationAccessor#Find %v id: %q\n", err, key.Encode())
		return nil, err
	}
	m.ID = key.Encode()
	return m, nil
}

func (aa *OrganizationAccessor) All(ctx context.Context) ([]*Organization, error) {
	q := datastore.NewQuery("Organizations")
	iter := q.Run(ctx)
	var res = []*Organization{}
	for {
		m := &Organization{}
		key, err := iter.Next(m)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		m.ID = key.Encode()
		res = append(res, m)
	}
	return res, nil
}
