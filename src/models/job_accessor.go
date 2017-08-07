package models

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type JobAccessor struct {
	Parent *Pipeline
}

var GlobalJobAccessor = &JobAccessor{}

var ErrNoSuchJob = errors.New("No such data in Jobs")

func (aa *JobAccessor) Find(ctx context.Context, id string) (*Job, error) {
	// log.Debugf(ctx, "JobAccessor#Find id: %q\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "JobAccessor#Find %v id: %q\n", err, id)
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
	// log.Debugf(ctx, "JobAccessor#Find key: %q\n", key)
	m := &Job{ID: id}
	err = datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchJob
	case err != nil:
		log.Errorf(ctx, "JobAccessor#Find %v id: %q\n", err, id)
		return nil, err
	}
	msg := &m.Message
	msg.EntriesToMap()
	return m, nil
}

func (aa *JobAccessor) Query() (*datastore.Query, error) {
	q := datastore.NewQuery("Jobs")
	if aa.Parent != nil {
		key, err := datastore.DecodeKey(aa.Parent.ID)
		if err != nil {
			return nil, err
		}
		q = q.Ancestor(key)
	}
	return q, nil
}

func (aa *JobAccessor) All(ctx context.Context) (Jobs, error) {
	q, err := aa.Query()
	if err != nil {
		return nil, err
	}
	iter := q.Run(ctx)
	var res = Jobs{}
	for {
		m := Job{}
		key, err := iter.Next(&m)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		m.ID = key.Encode()
		msg := &m.Message
		msg.EntriesToMap()
		res = append(res, &m)
	}
	return res, nil
}

func (aa *JobAccessor) WorkingAny(ctx context.Context) (bool, error) {
	cnt := 0
	for _, st := range WorkingJobStatuses {
		q, err := aa.Query()
		if err != nil {
			return false, err
		}
		q = q.Filter("Status =", st)
		c, err := q.Count(ctx)
		if err != nil {
			return false, err
		}
		cnt += c
	}
	return (cnt > 0), nil
}
