package models

import (
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
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

type InvalidOperation struct {
	Msg string
}

func (e *InvalidOperation) Error() string {
	return e.Msg
}

var processorFactory ProcessorFactory = &DefaultProcessorFactory{}

var ErrNoSuchPipeline = errors.New("No such data in Pipelines")

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
	}
)


func (pl *Pipeline) Destroy(ctx context.Context) error {
	if pl.Status != Closed {
		return &InvalidOperation{
			Msg: fmt.Sprintf("Can't destroy pipeline which is %v. Close before delete.", pl.Status),
		}
	}
	key, err := datastore.DecodeKey(pl.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (pl *Pipeline) Update(ctx context.Context) error {
	key, err := datastore.DecodeKey(pl.ID)
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, pl)
	if err != nil {
		return err
	}
	return nil
}

func (pl *Pipeline) Process(ctx context.Context, action string) error {
	processor, err := processorFactory.Create(ctx, action)
	if err != nil {
		return err
	}
	return processor.Process(ctx, pl)
}
