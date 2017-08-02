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

func (js JobStatus) String() string {
	r, ok := JobStatusToString[js]
	if !ok {
		return "<Invalid JobStatus>"
	}
	return r
}

func (js JobStatus) IsFinished() bool {
	switch js {
	case Failure, Success:
		return true
	default:
		return false
	}
}

const (
	Preparing JobStatus = iota
	Ready
	Publishing
	PublishError
	Published
	Executing
	Failure
	Success
)

var JobStatusToString = map[JobStatus]string{
	Preparing:    "Preparing",
	Ready:        "Ready",
	Publishing:   "Publishing",
	PublishError: "PublishError",
	Published:    "Published",
	Executing:    "Executing",
	Failure:      "Failure",
	Success:      "Success",
}

var (
	WorkingJobStatuses  = []JobStatus{Ready, Publishing, Published, Executing}
	LivingJobStatuses   = append([]JobStatus{Preparing}, WorkingJobStatuses...)
	FinishedJobStatuses = []JobStatus{PublishError, Failure, Success}
)

func (js JobStatus) Working() bool {
	return js.IncludedIn(WorkingJobStatuses)
}

func (js JobStatus) Living() bool {
	return js.IncludedIn(LivingJobStatuses)
}

func (js JobStatus) IncludedIn(statuses []JobStatus) bool {
	for _, st := range statuses {
		if js == st {
			return true
		}
	}
	return false
}

type (
	KeyValuePair struct {
		Name  string `datastore:"name"  validate:"required"`
		Value string `datastore:"value,noindex"`
	}

	JobMessage struct {
		AttributeMap     map[string]string `json:"attributes" datastore:"-"`
		AttributeEntries []KeyValuePair    `json:"-"          datastore:"attribute_entries"`
		Data             string            `json:"data" datastore:"data,noindex"`
	}
)

const JobIdKey = "concurrent_batch.job_id"

func (m *JobMessage) MapToEntries() {
	entries := []KeyValuePair{}
	for k, v := range m.AttributeMap {
		entries = append(entries, KeyValuePair{Name: k, Value: v})
	}
	m.AttributeEntries = entries
}

func (m *JobMessage) EntriesToMap() {
	kv := map[string]string{}
	for _, entry := range m.AttributeEntries {
		kv[entry.Name] = entry.Value
	}
	m.AttributeMap = kv
}

type (
	Job struct {
		ID         string     `json:"id"  datastore:"-"`
		Pipeline   *Pipeline  `json:"-"   validate:"required" datastore:"-"`
		IdByClient string     `json:"id_by_client" validate:"required" datastore:"id_by_client"`
		Status     JobStatus  `json:"status"       datastore:"status" `
		Message    JobMessage `json:"message" datastore:"message"`
		MessageID  string     `json:"message_id"   datastore:"message_id"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
	}
)

func (m *Job) Validate() error {
	v := validator.New()
	for k, val := range Validators {
		v.RegisterValidation(k, val)
	}
	err := v.Struct(m)
	return err
}

func (m *Job) Create(ctx context.Context) error {
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

	log.Debugf(ctx, "Job#Create: %v\n", m)

	err := m.Validate()
	if err != nil {
		return err
	}

	parentKey, err := datastore.DecodeKey(m.Pipeline.ID)
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Jobs", parentKey)
	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = res.Encode()
	return nil
}

func (m *Job) Update(ctx context.Context) error {
	if len(m.Message.AttributeEntries) == 0 {
		msg := &m.Message
		msg.MapToEntries()
	}

	m.UpdatedAt = time.Now()
	if m.Pipeline == nil {
		err := m.LoadPipeline(ctx)
		if err != nil {
			return err
		}
	}

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

func (m *Job) Destroy(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Job) LoadPipeline(ctx context.Context) error {
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
	if pl.Organization == nil {
		err = pl.LoadOrganization(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Job) JobMessage() *pubsub.PubsubMessage {
	entry := KeyValuePair{Name: JobIdKey, Value: m.ID}
	m.Message.AttributeEntries = append(m.Message.AttributeEntries, entry)
	if len(m.Message.AttributeMap) == 0 {
		msg := &m.Message
		msg.EntriesToMap()
	} else {
		m.Message.AttributeMap[JobIdKey] = m.ID
	}
	return &pubsub.PubsubMessage{
		Attributes: m.Message.AttributeMap,
		Data:       base64.StdEncoding.EncodeToString([]byte(m.Message.Data)),
	}
}

func (m *Job) Publish(ctx context.Context) (string, error) {
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

func (m *Job) PublishAndUpdate(ctx context.Context) error {
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

func (m *Job) CreateAndPublishIfPossible(ctx context.Context) error {
	return m.DoAndPublishIfPossible(ctx, m.Create)
}

func (m *Job) UpdateAndPublishIfPossible(ctx context.Context) error {
	return m.DoAndPublishIfPossible(ctx, m.Update)
}

func (m *Job) DoAndPublishIfPossible(ctx context.Context, f func(ctx context.Context) error) error {
	if m.Status == Ready {
		pl := m.Pipeline
		switch pl.Status {
		case Uninitialized, Waiting, Reserved, Building, Deploying:
			m.Status = Ready
		case Opened:
			m.Status = Publishing
		default:
			msg := fmt.Sprintf("Can't create and publish a job to a pipeline which is %v", pl.Status)
			return &InvalidOperation{Msg: msg}
		}
	}

	err := f(ctx)
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

func (m *Job) UpdateStatusIfGreaterThanBefore(ctx context.Context, completed bool, step JobStep, stepStatus JobStepStatus) error {
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
		case CANCELLING:
			newStatus = Failure
		case ACKSENDING:
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
