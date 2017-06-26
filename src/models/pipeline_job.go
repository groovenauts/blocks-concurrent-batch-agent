package models

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

type JobStatus int

const (
	Waiting JobStatus = iota
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
		IdByClient string        `json:"id_by_client" datastore:"id_by_client"`
		Status     JobStatus     `json:"status"       datastore:"status"`
		Message    PipelineJobMessage `json:"message" datastore:"message"`

	}
)

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
