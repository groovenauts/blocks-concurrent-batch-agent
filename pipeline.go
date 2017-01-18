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
	initialized Status = 0
	broken             = 1
	building           = 2
	opened             = 3
	resizing           = 4
	updating           = 5
	recreating         = 6
	closing            = 8
	closed             = 9
)

var processorFactory ProcessorFactory = &DefaultProcessorFactory{}

var ErrNoSuchPipeline = errors.New("No such data in Pipelines")

type (
	PipelineProps struct {
		ProjectID string `json:"project_id"`
		// JobTopicName							   string `json:"job_topic_name"`
		// JobSubscriptionName				   string `json:"job_subscription_name"`
		// JobSubscriptionAckDeadline	 int		`json:"job_subscription_ack_deadline"`
		// ProgressTopicName							   string `json:"progress_topic_name"`
		// ProgressSubscriptionName				 string `json:"progress_subscription_name"`
		// ProgressSubscriptionAckDeadline	 int		`json:"progress_subscription_ack_deadline"`
		// InstanceGroupName				 string `json:"instance_group_name"`
		// InstanceGroupSize				 int		`json:"instance_group_size"`
		// InstanceTemplateName		 string `json:"instance_template_name"`
		// StartupScript						 string `json:"startup_script"`
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
		log.Debugf(ctx, "GetAllPipeline pl: %v\n", pl)
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

func (pl *Pipeline) process(ctx context.Context, action string) error {
	processor, err := processorFactory.Create(ctx, action)
	if err != nil {
		return err
	}
	return processor.Process(ctx, pl)
}
