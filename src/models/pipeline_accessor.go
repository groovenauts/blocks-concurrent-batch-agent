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
	pl := &Pipeline{ID: id}
	err := pa.LoadByID(ctx, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (pa *PipelineAccessor) FindByKey(ctx context.Context, key *datastore.Key) (*Pipeline, error) {
	pl := &Pipeline{}
	err := pa.LoadByKey(ctx, key, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (pa *PipelineAccessor) LoadByID(ctx context.Context, pl *Pipeline) error {
	if pl.ID == "" {
		err := fmt.Errorf("No ID given to load a Pipeline %v", pl)
		log.Errorf(ctx, "Failed to load Pipeline because of %v\n", err)
		return err
	}

	key, err := datastore.DecodeKey(pl.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to decode id(%v) to key because of %v \n", pl.ID, err)
		return err
	}
	if pa.Parent != nil {
		parentKey, err := datastore.DecodeKey(pa.Parent.ID)
		if err != nil {
			return err
		}
		if !parentKey.Equal(key.Parent()) {
			return &InvalidParent{pl.ID}
		}
	}
	return pa.LoadByKey(ctx, key, pl)
}

func (pa *PipelineAccessor) LoadByKey(ctx context.Context, key *datastore.Key, pl *Pipeline) error {
	ctx = context.WithValue(ctx, "Pipeline.key", key)
	err := datastore.Get(ctx, key, pl)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return ErrNoSuchPipeline
	case err != nil:
		log.Errorf(ctx, "Failed to Get pipeline key(%v) to key because of %v \n", key, err)
		return err
	}
	pl.ID = key.Encode()
	return nil
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

func (pa *PipelineAccessor) WaitingQuery() (*datastore.Query, error) {
	q := datastore.NewQuery("Pipelines").Filter("Status =", Waiting)
	q, err := pa.considerParent(q)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (pa *PipelineAccessor) GetWaitings(ctx context.Context) ([]*Pipeline, error) {
	q, err := pa.WaitingQuery()
	if err != nil {
		return nil, err
	}
	return pa.GetByQuery(ctx, q.Order("CreatedAt"))
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

func (pa *PipelineAccessor) PendingsFor(ctx context.Context, jobIDs []string) ([]*Pipeline, error) {
	pipelines := map[string]*Pipeline{}
	for _, jobID := range jobIDs {
		q := datastore.NewQuery("Pipelines").Filter("Status =", Pending).Filter("Dependency.JobIDs =", jobID)
		r, err := pa.GetByQuery(ctx, q)
		if err != nil {
			return nil, err
		}
		for _, pl := range r {
			pipelines[pl.ID] = pl
		}
	}

	result := []*Pipeline{}
	for _, pl := range pipelines {
		result = append(result, pl)
	}
	return result, nil
}
