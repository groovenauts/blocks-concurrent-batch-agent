package models

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

// See https://godoc.org/google.golang.org/api/deploymentmanager/v2#OperationErrorErrors
//     https://godoc.org/google.golang.org/api/compute/v1#OperationErrorErrors
type OperationError struct {
	// Code: [Output Only] The error type identifier for this error.
	Code string `json:"code,omitempty"`

	// Location: [Output Only] Indicates the field in the request that
	// caused the error. This property is optional.
	Location string `json:"location,omitempty"`

	// Message: [Output Only] An optional, human-readable error message.
	Message string `json:"message,omitempty"`
}

type OperationLog struct {
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// See https://godoc.org/google.golang.org/api/deploymentmanager/v2#Operation
//     https://godoc.org/google.golang.org/api/compute/v1#Operation
type PipelineOperation struct {
	ID            string           `json:"id"                             datastore:"-"`
	Pipeline      *Pipeline        `json:"-"          validate:"required" datastore:"-"`
	ProjectID     string           `json:"project_id" validate:"required"`
	Zone          string           `json:"zone"				validate:"required"`
	Service       string           `json:"service"		validate:"required"`
	Name          string           `json:"name"       validate:"required"`
	OperationType string           `json:"operation_type" validate:"required"`
	Status        string           `json:"status"`
	Errors        []OperationError `json:"errors"`
	Logs          []OperationLog   `json:"logs"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (m *PipelineOperation) Validate() error {
	validator := validator.New()
	return validator.Struct(m)
}

func (m *PipelineOperation) Create(ctx context.Context) error {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = t
	}

	if m.Pipeline == nil {
		return fmt.Errorf("No pipeline to create PipelineOperation: %v\n", m)
	}
	parentKey, err := datastore.DecodeKey(m.Pipeline.ID)
	if err != nil {
		return err
	}
	key := datastore.NewIncompleteKey(ctx, "PipelineOperations", parentKey)

	resKey, err := m.ValidateAndPut(ctx, key)
	if err != nil {
		return err
	}
	m.ID = resKey.Encode()
	return nil
}

func (m *PipelineOperation) Update(ctx context.Context) error {
	m.UpdatedAt = time.Now()
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	_, err = m.ValidateAndPut(ctx, key)
	if err != nil {
		return err
	}
	return nil
}

func (m *PipelineOperation) ValidateAndPut(ctx context.Context, key *datastore.Key) (*datastore.Key, error) {
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

func (m *PipelineOperation) AppendLog(msg string) {
	m.Logs = append(m.Logs, OperationLog{CreatedAt: time.Now(), Message: msg})
}

func (m *PipelineOperation) ProcessDeploy(ctx context.Context, updater *DeploymentUpdater) error {
	return updater.Update(ctx, m,
		func() error {
			return m.LoadPipelineWith(ctx, func(pl *Pipeline) error {
				return pl.CompleteDeploying(ctx)
			})
		},
		func() error {
			return m.LoadPipelineWith(ctx, func(pl *Pipeline) error {
				return pl.FailDeploying(ctx)
			})
		},
	)
}

func (m *PipelineOperation) Done() bool {
	return m.Status == "DONE"
}

func (m *PipelineOperation) HasError() bool {
	return len(m.Errors) > 0
}

func (m *PipelineOperation) LoadPipeline(ctx context.Context) (*Pipeline, error) {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return nil, err
	}
	parentKey := key.Parent()
	pl, err := GlobalPipelineAccessor.FindByKey(ctx, parentKey)
	if err != nil {
		return nil, err
	}
	m.Pipeline = pl
	return pl, nil
}

func (m *PipelineOperation) LoadPipelineWith(ctx context.Context, f func(*Pipeline) error) error {
	pl, err := m.LoadPipeline(ctx)
	if err != nil {
		return err
	}
	return f(pl)
}
