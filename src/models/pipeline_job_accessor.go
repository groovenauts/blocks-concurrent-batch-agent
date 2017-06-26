package models

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type PipelineJobAccessor struct {
	Parent *Pipeline
}

var GlobalPipelineJobAccessor = &PipelineJobAccessor{}

var ErrNoSuchPipelineJob = errors.New("No such data in PipelineJobs")

func (aa *PipelineJobAccessor) Find(ctx context.Context, id string) (*PipelineJob, error) {
	// log.Debugf(ctx, "PipelineJobAccessor#Find id: %q\n", id)
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "PipelineJobAccessor#Find %v id: %q\n", err, id)
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
	// log.Debugf(ctx, "PipelineJobAccessor#Find key: %q\n", key)
	m := &PipelineJob{ID: id}
	err = datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchPipelineJob
	case err != nil:
		log.Errorf(ctx, "PipelineJobAccessor#Find %v id: %q\n", err, id)
		return nil, err
	}
	return m, nil
}


func (aa *PipelineJobAccessor) All(ctx context.Context) ([]*PipelineJob, error) {
	q := datastore.NewQuery("PipelineJobs")
	if aa.Parent != nil {
		key, err := datastore.DecodeKey(aa.Parent.ID)
		if err != nil {
			return nil, err
		}
		q = q.Ancestor(key)
	}
	iter := q.Run(ctx)
	var res = []*PipelineJob{}
	for {
		m := PipelineJob{}
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
