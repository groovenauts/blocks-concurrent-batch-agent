package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/net/context"
	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

type JobStatus int

func (js JobStatus) GreaterThan(other JobStatus) bool {
	return int(js) > int(other)
}

const (
	Waiting JobStatus = iota
	Publishing
	PublishError
	Published
	Executing
	Failure
	Success
)

var (
	WorkingJobStatuses  = []JobStatus{Waiting, Publishing, Published, Executing}
	FinishedJobStatuses = []JobStatus{PublishError, Failure, Success}
)

type (
	KeyValuePair struct {
		Name  string `datastore:"name"  validate:"required"`
		Value string `datastore:"value,noindex"`
	}

	PipelineJobMessage struct {
		AttributeMap     map[string]string `json:"attributes" datastore:"-"`
		AttributeEntries []KeyValuePair    `json:"-"          datastore:"attribute_entries"`
		Data             string            `json:"data" datastore:"data,noindex"`
	}
)

const PipelineJobIdKey = "concurrent_batch.pipeline_job_id"

func (m *PipelineJobMessage) MapToEntries() {
	entries := []KeyValuePair{}
	for k, v := range m.AttributeMap {
		entries = append(entries, KeyValuePair{Name: k, Value: v})
	}
	m.AttributeEntries = entries
}

func (m *PipelineJobMessage) EntriesToMap() {
	kv := map[string]string{}
	for _, entry := range m.AttributeEntries {
		kv[entry.Name] = entry.Value
	}
	m.AttributeMap = kv
}

type (
	PipelineJob struct {
		ID         string             `json:"id"  datastore:"-"`
		Pipeline   *Pipeline          `json:"-"   validate:"required" datastore:"-"`
		IdByClient string             `json:"id_by_client" validate:"required" datastore:"id_by_client"`
		Status     JobStatus          `json:"status"       datastore:"status" `
		Message    PipelineJobMessage `json:"message" datastore:"message"`
		MessageID  string             `json:"message_id"   datastore:"message_id"`
		CreatedAt  time.Time          `json:"created_at"`
		UpdatedAt  time.Time          `json:"updated_at"`
	}
)

func (m *PipelineJob) Validate() error {
	v := validator.New()
	for k, val := range Validators {
		v.RegisterValidation(k, val)
	}
	err := v.Struct(m)
	return err
}

func (m *PipelineJob) Create(ctx context.Context) error {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = t
	}

	if len(m.Message.AttributeEntries) == 0 {
		msg := &m.Message
		msg.MapToEntries()
	}

	log.Debugf(ctx, "PipelineJob#Create: %v\n", m)

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
	if len(m.Message.AttributeEntries) == 0 {
		msg := &m.Message
		msg.MapToEntries()
	}

	m.UpdatedAt = time.Now()

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

func (m *PipelineJob) JobMessage() *pubsub.PubsubMessage {
	entry := KeyValuePair{Name: PipelineJobIdKey, Value: m.ID}
	m.Message.AttributeEntries = append(m.Message.AttributeEntries, entry)
	if len(m.Message.AttributeMap) == 0 {
		msg := &m.Message
		msg.EntriesToMap()
	} else {
		m.Message.AttributeMap[PipelineJobIdKey] = m.ID
	}
	return &pubsub.PubsubMessage{
		Attributes: m.Message.AttributeMap,
		Data:       base64.StdEncoding.EncodeToString([]byte(m.Message.Data)),
	}
}

func (m *PipelineJob) Publish(ctx context.Context) (string, error) {
	msg := m.JobMessage()
	log.Debugf(ctx, "m.JobMessage: %v\n", msg)
	topic := m.Pipeline.JobTopicFqn()
	log.Debugf(ctx, "Sending message to %v: %v\n", topic, msg)

	req := &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{msg},
	}

	msgId, err := GlobalPublisher.Publish(ctx, topic, req)
	if err != nil {
		return "", err
	}

	m.MessageID = msgId
	err = m.Update(ctx)
	if err != nil {
		return "", err
	}

	return msgId, nil
}

func (m *PipelineJob) PublishAndUpdate(ctx context.Context) error {
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

func (m *PipelineJob) CreateAndPublishIfPossible(ctx context.Context) error {
	pl := m.Pipeline
	switch pl.Status {
	case Uninitialized, Pending, Reserved, Building, Deploying:
		m.Status = Waiting
	case Opened:
		m.Status = Publishing
	default:
		msg := fmt.Sprintf("Can't create and publish a job to a pipeline which is %v", pl.Status)
		return &InvalidOperation{Msg: msg}
	}

	err := m.Create(ctx)
	if err != nil {
		return err
	}

	if m.Status != Publishing {
		return nil
	}

	err = m.PublishAndUpdate(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *PipelineJob) UpdateStatusIfGreaterThanBefore(ctx context.Context, completed bool, step JobStep, stepStatus JobStepStatus) error {
	if completed {
		m.Status = Success
		err := m.Update(ctx)
		if err != nil {
			return err
		}
		return nil
	}

	newStatus := m.Status
	switch stepStatus {
	case STARTING:
		// Ignore
	case SUCCESS:
		switch step {
		case INITIALIZING, DOWNLOADING, EXECUTING, UPLOADING, NACKSENDING:
			newStatus = Executing
		case CLEANUP:
			// Do nothing
		case CANCELLING  :
			newStatus = Failure
		case ACKSENDING  :
			newStatus = Success
		}
	case FAILURE:
		switch step {
		case INITIALIZING, DOWNLOADING, EXECUTING, UPLOADING:
			newStatus = Executing
		}
	}

	if newStatus.GreaterThan(m.Status) {
		m.Status = newStatus
		err := m.Update(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
