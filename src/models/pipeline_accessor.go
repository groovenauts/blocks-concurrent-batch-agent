package models

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

func CreatePipeline(ctx context.Context, pl *Pipeline) error {
	validator := validator.New()
	err := validator.Struct(pl)
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Pipelines", nil)
	res, err := datastore.Put(ctx, key, pl)
	if err != nil {
		return err
	}
	pl.ID = res.Encode()
	return nil
}

func FindPipeline(ctx context.Context, id string) (*Pipeline, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "Failed to decode id(%v) to key because of %v \n", id, err)
		return nil, err
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

func GetAllPipelines(ctx context.Context) ([]*Pipeline, error) {
	q := datastore.NewQuery("Pipelines")
	return GetPipelinesByQuery(ctx, q)
}

func GetPipelinesByStatus(ctx context.Context, st Status) ([]*Pipeline, error) {
	q := datastore.NewQuery("Pipelines").Filter("Status =", st)
	return GetPipelinesByQuery(ctx, q)
}

func GetPipelinesByQuery(ctx context.Context, q *datastore.Query) ([]*Pipeline, error) {
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

func GetPipelineIDsByStatus(ctx context.Context, st Status) ([]string, error) {
	q := datastore.NewQuery("Pipelines").Filter("Status =", st)
	return GetPipelineIDsByQuery(ctx, q)
}

func GetPipelineIDsByQuery(ctx context.Context, q *datastore.Query) ([]string, error) {
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

func GetActiveSubscriptions(ctx context.Context) ([]*Subscription, error) {
	r := []*Subscription{}
	pipelines, err := GetPipelinesByStatus(ctx, Opened)
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
