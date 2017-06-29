package models

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

// Status constants
type Status int

const (
	Initialized Status = iota
	Broken
	Building
	Deploying
	Opened
	Closing
	Closing_error
	Closed
)

var StatusStrings = map[Status]string{
	Initialized:   "initialized",
	Broken:        "broken",
	Building:      "building",
	Deploying:     "deploying",
	Opened:        "opened",
	Closing:       "closing",
	Closing_error: "closing_error",
	Closed:        "closed",
}

func (st Status) String() string {
	res, ok := StatusStrings[st]
	if !ok {
		return "Invalid Status: " + strconv.Itoa(int(st))
	}
	return res
}

var processorFactory ProcessorFactory = &DefaultProcessorFactory{}

type (
	// See https://godoc.org/google.golang.org/api/deploymentmanager/v2#OperationErrorErrors
	DeploymentError struct {
		// Code: [Output Only] The error type identifier for this error.
		Code string `json:"code,omitempty"`

		// Location: [Output Only] Indicates the field in the request that
		// caused the error. This property is optional.
		Location string `json:"location,omitempty"`

		// Message: [Output Only] An optional, human-readable error message.
		Message string `json:"message,omitempty"`
	}

	PipelineVmDisk struct {
		// DiskName    string `json:"disk_name,omitempty"` // Don't support diskName to keep simple naming rule
		DiskSizeGb  int    `json:"disk_size_gb,omitempty"`
		DiskType    string `json:"disk_type,omitempty"`
		SourceImage string `json:"source_image" validate:"required"`
	}

	Pipeline struct {
		ID                     string            `json:"id"             datastore:"-"`
		Organization           *Organization     `json:"-"              validate:"required" datastore:"-"`
		Name                   string            `json:"name"           validate:"required"`
		ProjectID              string            `json:"project_id"     validate:"required"`
		Zone                   string            `json:"zone"           validate:"required"`
		BootDisk               PipelineVmDisk    `json:"boot_disk"`
		MachineType            string            `json:"machine_type"   validate:"required"`
		Preemptible            bool              `json:"preemptible,omitempty"`
		StackdriverAgent       bool              `json:"stackdriver_agent,omitempty"`
		TargetSize             int               `json:"target_size"    validate:"required"`
		ContainerSize          int               `json:"container_size" validate:"required"`
		ContainerName          string            `json:"container_name" validate:"required"`
		Command                string            `json:"command"` // allow blank
		Status                 Status            `json:"status"`
		Dryrun                 bool              `json:"dryrun"`
		DeploymentName         string            `json:"deployment_name"`
		DeployingOperationName string            `json:"deploying_operation_name"`
		DeployingErrors        []DeploymentError `json:"deploying_errors"`
		ClosingOperationName   string            `json:"closing_operation_name"`
		ClosingErrors          []DeploymentError `json:"closing_errors"`
		TokenConsumption       int               `json:"token_consumption"`
	}
)

func (m *Pipeline) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Pipeline) Create(ctx context.Context) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	parentKey, err := datastore.DecodeKey(m.Organization.ID)
	if err != nil {
		return err
	}

	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
		if err != nil {
			return err
		}
		newAmount := org.TokenAmount - m.TokenConsumption
		if newAmount < 0 {
			msg := fmt.Sprintf("Insufficient tokens; %v has only %v tokens but %v required %v tokens", org.Name, org.TokenAmount, m.Name, m.TokenConsumption)
			return &InvalidOperation{Msg: msg}
		}
		org.TokenAmount = newAmount
		err = org.Update(ctx)
		if err != nil {
			return err
		}

		key := datastore.NewIncompleteKey(ctx, "Pipelines", parentKey)
		res, err := datastore.Put(ctx, key, m)
		if err != nil {
			return err
		}
		m.ID = res.Encode()

		return nil
	}, nil)

	if err != nil {
		log.Errorf(ctx, "Transaction failed: %v\n", err)
		return err
	}

	return nil
}

func (m *Pipeline) Destroy(ctx context.Context) error {
	if m.Status != Closed {
		return &InvalidOperation{
			Msg: fmt.Sprintf("Can't destroy pipeline which is %v. Close before delete.", m.Status),
		}
	}
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) Update(ctx context.Context) error {
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
	return nil
}

func (m *Pipeline) Process(ctx context.Context, action string) error {
	processor, err := processorFactory.Create(ctx, action)
	if err != nil {
		return err
	}
	return processor.Process(ctx, m)
}

func (m *Pipeline) CompleteClosing(ctx context.Context) error {
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
		if err != nil {
			return err
		}
		org.TokenAmount = org.TokenAmount + m.TokenConsumption
		err = org.Update(ctx)
		if err != nil {
			return err
		}

		m.Status = Closed
		return m.Update(ctx)
	}, nil)
	return err
}

func (m *Pipeline) LoadOrganization(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to decode Key of pipeline %v because of %v\n", m.ID, err)
		return err
	}
	orgKey := key.Parent()
	if orgKey == nil {
		log.Errorf(ctx, "Pipline key has no parent. ID: %v\n", m.ID)
		panic("Invalid pipeline key")
	}
	org, err := GlobalOrganizationAccessor.FindByKey(ctx, orgKey)
	if err != nil {
		return err
	}
	m.Organization = org
	return nil
}

func (m *Pipeline) JobAccessor() *PipelineJobAccessor {
	return &PipelineJobAccessor{Parent: m}
}

func (m *Pipeline) Reload(ctx context.Context) error {
	err := GlobalPipelineAccessor.LoadByID(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) WaitUntil(ctx context.Context, st Status, interval, timeout time.Duration) error {
	t0 := time.Now()
	deadline := t0.Add(timeout)
	for deadline.After(time.Now()) {
		m.Reload(ctx)
		if m.Status == st {
			return nil
		}
		time.Sleep(interval)
	}
	return ErrTimeout
}

func (m *Pipeline) JobTopicName() string {
	return fmt.Sprintf("%s-job-topic", m.Name)
}

func (m *Pipeline) JobTopicFqn() string {
	return fmt.Sprintf("projects/%s/topics/%s", m.ProjectID, m.JobTopicName())
}
