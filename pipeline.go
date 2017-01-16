package pipeline

import (
	"golang.org/x/net/context"
)

type (
	Pipeline struct {
		ProjectID								 string `json:"project_id"`
		TopicName								 string `json:"topic_name"`
		SubscriptionName				 string `json:"subscription_name"`
		SubscriptionAckDeadline	 int		`json:"subscription_ack_deadline"`
		InstanceGroupName				 string `json:"instance_group_name"`
		InstanceGroupSize				 int		`json:"instance_group_size"`
		InstanceTemplateName		 string `json:"instance_template_name"`
		StartupScript						 string `json:"startup_script"`
		Status									 int		`json:"status"`
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
