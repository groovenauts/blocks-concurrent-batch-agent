package models

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

type PipelineInstanceSizeLog struct {
	ID        string    `json:"id"                        datastore:"-"`
	pipeline  *Pipeline `            validate:"required"`
	Size      int       `json:"size" validate:"required"`
	CreatedAt time.Time `json:"time" validate:"required"`
}

func (m *PipelineInstanceSizeLog) Validate() error {
	validator := validator.New()
	return validator.Struct(m)
}

func (m *PipelineInstanceSizeLog) Create(ctx context.Context) error {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}

	if m.pipeline == nil {
		return fmt.Errorf("No pipeline to create PipelineInstanceSizeLog: %v\n", m)
	}
	parentKey, err := datastore.DecodeKey(m.pipeline.ID)
	if err != nil {
		return err
	}
	key := datastore.NewIncompleteKey(ctx, "PipelineInstanceSizeLogs", parentKey)

	resKey, err := m.ValidateAndPut(ctx, key)
	if err != nil {
		return err
	}
	m.ID = resKey.Encode()
	return nil
}

func (m *PipelineInstanceSizeLog) ValidateAndPut(ctx context.Context, key *datastore.Key) (*datastore.Key, error) {
	err := m.Validate()
	if err != nil {
		return nil, err
	}

	key, err = datastore.Put(ctx, key, m)
	if err != nil {
		return nil, err
	}
	return key, nil
}
