package pipeline

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Status constants
type Status int

const (
	initialized Status = iota
	broken
	building
	opened
	closing
	closed
	resizing
	updating
	recreating
)

type (
	Pipeline struct {
		ProjectID								     string `json:"project_id"`
		JobTopicName							   string `json:"job_topic_name"`
		JobSubscriptionName				   string `json:"job_subscription_name"`
		JobSubscriptionAckDeadline	 int		`json:"job_subscription_ack_deadline"`
		ProgressTopicName							   string `json:"progress_topic_name"`
		ProgressSubscriptionName				 string `json:"progress_subscription_name"`
		ProgressSubscriptionAckDeadline	 int		`json:"progress_subscription_ack_deadline"`
		InstanceGroupName				 string `json:"instance_group_name"`
		InstanceGroupSize				 int		`json:"instance_group_size"`
		InstanceTemplateName		 string `json:"instance_template_name"`
		StartupScript						 string `json:"startup_script"`
		Status									 Status		`json:"status"`
	}
)

func CreatePipeline(ctx context.Context, pl *Pipeline) (string, error) {
	key := datastore.NewIncompleteKey(ctx, "Pipelines", nil)
	res, err := datastore.Put(ctx, key, pl)
	if err != nil {
		return "", err
	}
	ctx = context.WithValue(ctx, "Pipeline.key", res)
	return res.Encode(), nil
}

func FindPipeline(ctx context.Context, id string) (*Pipeline, error) {
	key, err := datastore.DecodeKey(id)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "Pipeline.key", key)
	pl := &Pipeline{}
	if err := datastore.Get(ctx, key, pl); err != nil {
		return nil, err
	}
	return pl, nil
}

func GetAllPipeline(ctx context.Context) ([]Pipeline, error) {
	q := datastore.NewQuery("Pipelines")
	var res []Pipeline
	_, err := q.GetAll(ctx, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetAllActivePipelineIDs(ctx context.Context) ([]string, error) {
	q := datastore.NewQuery("Pipelines").Filter("status !=", closed).KeysOnly()
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
	if pl.Status != closed {
		return fmt.Errorf("Can't destroy pipeline which has status: %v", pl.Status)
	}
	key := ctx.Value("Pipeline.key").(*datastore.Key)
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}
