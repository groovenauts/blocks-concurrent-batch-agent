package models

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type PipelineAccessor struct {
	Parent *Organization
}

var GlobalPipelineAccessor = &PipelineAccessor{}

var ErrNoSuchPipeline = errors.New("No such data in Pipelines")

func (pa *PipelineAccessor) Find(ctx context.Context, id string) (*Pipeline, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "Failed to decode id(%v) to key because of %v \n", id, err)
		return nil, err
	}
	if pa.Parent != nil {
		parentKey, err := datastore.DecodeKey(pa.Parent.ID)
		if err != nil {
			return nil, err
		}
		if !parentKey.Equal(key.Parent()) {
			return nil, &InvalidParent{id}
		}
	}
	ctx = context.WithValue(ctx, "Pipeline.key", key)
	pl := &Pipeline{ID: id}
	err = datastore.Get(ctx, key, pl)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchPipeline
	case err != nil:
		log.Errorf(ctx, "Failed to Get pipeline key(%v) to key because of %v \n", key, err)
		return nil, err
	}
	return pl, nil
}

func (pa *PipelineAccessor) considerParent(q *datastore.Query) (*datastore.Query, error) {
	if pa.Parent == nil {
		return q, nil
	}
	key, err := datastore.DecodeKey(pa.Parent.ID)
	if err != nil {
		return nil, err
	}
	return q.Ancestor(key), nil
}

func (pa *PipelineAccessor) GetAll(ctx context.Context) ([]*Pipeline, error) {
	q := datastore.NewQuery("Pipelines")
	return pa.GetByQuery(ctx, q)
}

func (pa *PipelineAccessor) GetByStatus(ctx context.Context, st Status) ([]*Pipeline, error) {
	q := datastore.NewQuery("Pipelines").Filter("Status =", st)
	return pa.GetByQuery(ctx, q)
}

func (pa *PipelineAccessor) GetByQuery(ctx context.Context, q *datastore.Query) ([]*Pipeline, error) {
	q, err := pa.considerParent(q)
	if err != nil {
		return nil, err
	}
	iter := q.Run(ctx)
	var res = []*Pipeline{}
	for {
		pl := Pipeline{}
		key, err := iter.Next(&pl)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		pl.ID = key.Encode()
		res = append(res, &pl)
	}
	return res, nil
}

func (pa *PipelineAccessor) GetIDsByStatus(ctx context.Context, st Status) ([]string, error) {
	q := datastore.NewQuery("Pipelines").Filter("Status =", st)
	q, err := pa.considerParent(q)
	if err != nil {
		return nil, err
	}
	return pa.GetIDsByQuery(ctx, q)
}

func (pa *PipelineAccessor) GetIDsByQuery(ctx context.Context, q *datastore.Query) ([]string, error) {
	q, err := pa.considerParent(q)
	if err != nil {
		return nil, err
	}
	keys, err := q.KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, key := range keys {
		res = append(res, key.Encode())
	}
	return res, nil
}

type Subscription struct {
	PipelineID string `json:"pipeline_id"`
	Pipeline   string `json:"pipeline"`
	Name       string `json:"subscription"`
}

func (pa *PipelineAccessor) GetActiveSubscriptions(ctx context.Context) ([]*Subscription, error) {
	r := []*Subscription{}
	pipelines, err := pa.GetByStatus(ctx, Opened)
	if err != nil {
		return nil, err
	}
	for _, pipeline := range pipelines {
		r = append(r, &Subscription{
			PipelineID: pipeline.ID,
			Pipeline:   pipeline.Name,
			Name:       fmt.Sprintf("projects/%v/subscriptions/%v-progress-subscription", pipeline.ProjectID, pipeline.Name),
		})
	}
	return r, nil
}
