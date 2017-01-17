package pipeline

import (
	"golang.org/x/net/context"
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
	return "encoded key", nil
}

func FindPipeline(ctx context.Context, id string) (*Pipeline, error) {
	return &Pipeline{}, nil
}

func GetAllPipeline(ctx context.Context) ([]Pipeline, error) {
	return []Pipeline{ Pipeline{} }, nil
}

func GetAllActivePipelineIDs(ctx context.Context) ([]string, error) {
	return []string{""}, nil
}

func (pl *Pipeline) destroy(ctx context.Context) error {
	return nil
}
