package pipeline

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

// Status constants
type Status int

const (
	initialized Status = iota
	broken
	building
	deploying
	opened
	closing
	closed
)

var processorFactory ProcessorFactory = &DefaultProcessorFactory{}

var ErrNoSuchPipeline = errors.New("No such data in Pipelines")

type (
	// See https://godoc.org/google.golang.org/api/deploymentmanager/v2#OperationErrorErrors
	DeploymentError struct {
		// Code: [Output Only] The error type identifier for this error.
		Code string `json:"code,omitempty"`

		// Location: [Output Only] Indicates the field in the request that
		// caused the error. This property is optional.
		Location string `json:"location,omitempty"`

		// Message: [Output Only] An optional, human-readable error message.
		Message string `json:"message,omitempty"`
	}

	PipelineProps struct {
		Name           string            `json:"name"           validate:"required"`
		ProjectID      string            `json:"project_id"     validate:"required"`
		Zone           string            `json:"zone"           validate:"required"`
		SourceImage    string            `json:"source_image"   validate:"required"`
		MachineType    string            `json:"machine_type"   validate:"required"`
		TargetSize     int               `json:"target_size"    validate:"required"`
		ContainerSize  int               `json:"container_size" validate:"required"`
		ContainerName  string            `json:"container_name" validate:"required"`
		Command        string            `json:"command"        validate:"required"`
		Status         Status            `json:"status"`
		Dryrun         bool              `json:"dryrun"`
		DeploymentName string            `json:"deployment_name"`
		OperationName  string            `json:"operation_name"`
		Errors         []DeploymentError `json:"errors"`
	}

	Pipeline struct {
		ID    string        `json:"id"`
		Props PipelineProps `json:"props"`
	}
)

func CreatePipeline(ctx context.Context, plp *PipelineProps) (*Pipeline, error) {
	validator := validator.New()
	err := validator.Struct(plp)
	if err != nil {
		return nil, err
	}

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
		log.Errorf(ctx, "@FindPipeline %v id: %v\n", err, id)
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
		key, err := iter.Next(&pl.Props)
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
