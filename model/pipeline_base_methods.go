package model

import (
	"fmt"
)

func (m *PipelineBase) JobTopicName() string {
	return fmt.Sprintf("%s-job-topic", m.Name)
}

func (m *PipelineBase) JobTopicFqn() string {
	return fmt.Sprintf("projects/%s/topics/%s", m.ProjectID, m.JobTopicName())
}

func (m *PipelineBase) ProgressSubscriptionName() string {
	return fmt.Sprintf("%s-progress-subscription", m.Name)
}

func (m *PipelineBase) ProgressSubscriptionFqn() string {
	return fmt.Sprintf("projects/%s/subscriptions/%s", m.ProjectID, m.ProgressSubscriptionName())
}
