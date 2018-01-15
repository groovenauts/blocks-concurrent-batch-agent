package models

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type PipelineOperationAccessor struct {
	Parent *Pipeline
}

var GlobalPipelineOperationAccessor = &PipelineOperationAccessor{}

var ErrNoSuchPipelineOperation = errors.New("No such data in PipelineOperations")

func (aa *PipelineOperationAccessor) Find(ctx context.Context, id string) (*PipelineOperation, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "Failed to DecodeKey at PipelineOperationAccessor#Find %v id: %q\n", err, id)
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
	m := &PipelineOperation{}
	err = datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchPipelineOperation
	case err != nil:
		log.Errorf(ctx, "Failed to Get at PipelineOperationAccessor#Find %v id: %q\n", err, id)
		return nil, err
	}
	m.ID = key.Encode()
	return m, nil
}

func (aa *PipelineOperationAccessor) Query() (*datastore.Query, error) {
	q := datastore.NewQuery("PipelineOperations")
	if aa.Parent != nil {
		key, err := datastore.DecodeKey(aa.Parent.ID)
		if err != nil {
			return nil, err
		}
		q = q.Ancestor(key)
	}
	return q, nil
}

func (aa *PipelineOperationAccessor) All(ctx context.Context) ([]*PipelineOperation, error) {
	q, err := aa.Query()
	if err != nil {
		return nil, err
	}
	iter := q.Run(ctx)
	var res = []*PipelineOperation{}
	for {
		m := PipelineOperation{}
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
