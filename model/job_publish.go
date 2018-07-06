package model

import (
	"encoding/base64"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"

	pubsub "google.golang.org/api/pubsub/v1"
)

func (m *Job) Publish(ctx context.Context) error {
	msg := m.BuildMessage()

	pbStore := &PipelineBaseStore{}
	pb, err := pbStore.Get(ctx, m.Parent.StringID())
	if err != nil {
		return err
	}

	topic := pb.JobTopicFqn()
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

const JobIdKey = "concurrent_batch.job_id"

func (m *Job) BuildMessage() *pubsub.PubsubMessage {
	entry := JobKeyValuePair{Name: JobIdKey, Value: m.Id}
	m.Message.AttributeEntries = append(m.Message.AttributeEntries, entry)
	return &pubsub.PubsubMessage{
		Attributes: m.Message.AttributeEntries.Map(),
		Data:       base64.StdEncoding.EncodeToString([]byte(m.Message.Data)),
	}
}



func (pairs JobKeyValuePairs) Map() map[string]string {
	r := map[string]string{}
	for _, entry := range pairs {
		r[entry.Name] = entry.Value
	}
	return r
}
