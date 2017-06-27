package models

import (
	"encoding/base64"
	"encoding/json"

	"golang.org/x/net/context"
	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"gopkg.in/go-playground/validator.v9"
)

type JobStatus int

const (
	Waiting JobStatus = iota
	Publishing
	PublishError
	Published
	// Failure
	// Success
)

type (
	PipelineJobMessage struct {
		AttributesJson string    `json:"attributes_json" datastore:"attributes_json"`
		Data           string    `json:"data" datastore:"data"`
	}

	PipelineJob struct {
		ID         string        `json:"id"  datastore:"-"`
		Pipeline   *Pipeline     `json:"-"   validate:"required" datastore:"-"`
		IdByClient string        `json:"id_by_client" validate:"required" datastore:"id_by_client"`
		Status     JobStatus     `json:"status"       datastore:"status" `
		Message    PipelineJobMessage `json:"message" datastore:"message"`
		MessageID  string        `json:"message_id"   datastore:"message_id"`
	}
)

func (m *PipelineJob) Attributes() (map[string]string, error) {
	var result map[string]string
	err := json.Unmarshal([]byte(m.Message.AttributesJson), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *PipelineJob) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *PipelineJob) Create(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	parentKey, err := datastore.DecodeKey(m.Pipeline.ID)
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "PipelineJobs", parentKey)
	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = res.Encode()
	return nil
}

func (m *PipelineJob) Update(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}

	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = key.Encode()
	return nil
}

func (m *PipelineJob) Destroy(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *PipelineJob) LoadPipeline(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to decode Key of pipeline %v because of %v\n", m.ID, err)
		return err
	}
	plKey := key.Parent()
	if plKey == nil {
		log.Errorf(ctx, "Pipline key has no parent. ID: %v\n", m.ID)
		panic("Invalid pipeline key")
	}
	pl, err := GlobalPipelineAccessor.FindByKey(ctx, plKey)
	if err != nil {
		return err
	}
	m.Pipeline = pl
	return nil
}

func (m *PipelineJob) Publish(ctx context.Context) (string, error) {
	// https://cloud.google.com/appengine/docs/standard/go/issue-requests
	client := urlfetch.Client(ctx)

	service, err := pubsub.New(client)
	if err != nil {
		log.Criticalf(ctx, "Failed to create pubsub.Service: %v\n", err)
		return "", err
	}

	attrs, err := m.Attributes()
	if err != nil {
		return "", err
	}

	call := service.Projects.Topics.Publish(m.Pipeline.JobTopicFqn(), &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{
			&pubsub.PubsubMessage{
				Attributes: attrs,
				Data:       base64.StdEncoding.EncodeToString([]byte(m.Message.Data)),
			},
		},
	})
	res, err := call.Do()
	if err != nil {
		log.Errorf(ctx, "Publish error: %v\n", err)
		return "", err
	}

	return res.MessageIds[0], nil
}

func (m *PipelineJob) PublishAndUpdate(ctx context.Context) error {
	m.Status = Publishing
	err := m.Update(ctx)
	if err != nil {
		return err
	}

	msgId, err := m.Publish(ctx)
	if err != nil {
		m.Status = PublishError
		e2 := m.Update(ctx)
		if e2 != nil {
			return e2
		}
		return err
	}

	m.MessageID = msgId
	m.Status = Published
	err = m.Update(ctx)
	if err != nil {
		return err
	}

	return nil
}
