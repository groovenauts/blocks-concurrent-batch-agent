package models

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type OrganizationAccessor struct {
}

var GlobalOrganizationAccessor = &OrganizationAccessor{}

var ErrNoSuchOrganization = errors.New("No such data in Organizations")

func (aa *OrganizationAccessor) Find(ctx context.Context, id string) (*Organization, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "OrganizationAccessor#Find %v id: %q\n", err, id)
		return nil, err
	}
	ctx = context.WithValue(ctx, "Organization.key", key)
	m := &Organization{}
	err = datastore.Get(ctx, key, m)
	m.ID = id
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchOrganization
	case err != nil:
		log.Errorf(ctx, "OrganizationAccessor#Find %v id: %q\n", err, id)
		return nil, err
	}
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
