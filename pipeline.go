package pipeline

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
