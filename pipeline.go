package pipeline

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// Status constants
type Status int

const (
	initialized Status = iota
	broken
	building
	opened
	resizing
	updating
	recreating
	closing
	closed
)

var processorFactory ProcessorFactory = &DefaultProcessorFactory{}

var ErrNoSuchPipeline = errors.New("No such data in Pipelines")

type (
	PipelineProps struct {
		Name      string `json:"name"`
		ProjectID string `json:"project_id"`
		Zone string `json:"zone"`
		SourceImage string `json:"source_image"`
		MachineType string `json:"machine_type"`
		TargetSize int `json:"target_size"`
		ContainerSize int `json:"container_size"`
		ContainerName string `json:"container_name"`
		Command string `json:"command"`
		Status Status `json:"status"`
	}

	Pipeline struct {
		ID    string        `json:"id"`
		Props PipelineProps `json:"props"`
	}
)

func CreatePipeline(ctx context.Context, plp *PipelineProps) (*Pipeline, error) {
	key := datastore.NewIncompleteKey(ctx, "Pipelines", nil)
	res, err := datastore.Put(ctx, key, plp)
	if err != nil {
		return nil, err
	}
	return &Pipeline{ID: res.Encode(), Props: *plp}, nil
}

func FindPipeline(ctx context.Context, id string) (*Pipeline, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		log.Errorf(ctx, "@FindPipeline %v id: %v\n", err, id)
		return nil, err
	}
	ctx = context.WithValue(ctx, "Pipeline.key", key)
	pl := &Pipeline{ID: id}
	err = datastore.Get(ctx, key, &pl.Props)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return nil, ErrNoSuchPipeline
	case err != nil:
		log.Errorf(ctx, "@withPipeline %v id: %v\n", err, id)
		return nil, err
	}
	return pl, nil
}

func GetAllPipeline(ctx context.Context) ([]Pipeline, error) {
	q := datastore.NewQuery("Pipelines")
	iter := q.Run(ctx)
	var res = []Pipeline{}
	for {
		pl := Pipeline{}
		key, err := iter.Next(&pl.Props)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		pl.ID = key.Encode()
		res = append(res, pl)
	}
	return res, nil
}

func GetAllActivePipelineIDs(ctx context.Context) ([]string, error) {
	q := datastore.NewQuery("Pipelines").Filter("Status <", closed).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, key := range keys {
		res = append(res, key.Encode())
	}
	return res, nil
}

func (pl *Pipeline) destroy(ctx context.Context) error {
	plp := pl.Props
	if plp.Status != closed {
		return fmt.Errorf("Can't destroy pipeline which has status: %v", plp.Status)
	}
	key, err := datastore.DecodeKey(pl.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (pl *Pipeline) update(ctx context.Context) error {
	key, err := datastore.DecodeKey(pl.ID)
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, &pl.Props)
	if err != nil {
		return err
	}
	return nil
}

func (pl *Pipeline) process(ctx context.Context, action string) error {
	processor, err := processorFactory.Create(ctx, action)
	if err != nil {
		return err
	}
	return processor.Process(ctx, pl)
}
