package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func (job *Job) Publish(ctx context.Context) error {
	msg := m.BuildMessage()
	topic := m.PipelineBase.JobTopicFqn()
	log.Debugf(ctx, "Job.Publish msg %v to topic %q\n", msg, topic)

	req := &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{msg},
	}

	msgId, err := GlobalPublisher.Publish(ctx, topic, req)
	if err != nil {
		return err
	}

	m.MessageID = msgId
	m.PublishedAt = time.Now()
	m.Status = Published

	return nil
}

func (job *Job) BuildMessage(ctx context.Context) error {
}
