package models

import (
	"errors"
	"fmt"

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
	m := &PipelineOperation{ID: id}
	err := aa.LoadByID(ctx, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (pa *PipelineOperationAccessor) LoadByID(ctx context.Context, m *PipelineOperation) error {
	if m.ID == "" {
		err := fmt.Errorf("No ID given to load a Pipeline %v", m)
		log.Errorf(ctx, "Failed to load PipelineOperation because of %v\n", err)
		return err
	}

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to decode id(%v) to key because of %v \n", m.ID, err)
		return err
	}
	if pa.Parent != nil {
		parentKey, err := datastore.DecodeKey(pa.Parent.ID)
		if err != nil {
			return err
		}
		if !parentKey.Equal(key.Parent()) {
			return &InvalidParent{m.ID}
		}
	}
	return pa.LoadByKey(ctx, key, m)
}

func (pa *PipelineOperationAccessor) LoadByKey(ctx context.Context, key *datastore.Key, m *PipelineOperation) error {
	ctx = context.WithValue(ctx, "Pipeline.key", key)
	err := datastore.Get(ctx, key, m)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return ErrNoSuchPipelineOperation
	case err != nil:
		log.Errorf(ctx, "Failed to Get PipelineOperation key(%v) to key because of %v \n", key, err)
		return err
	}
	m.ID = key.Encode()
	return nil
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
