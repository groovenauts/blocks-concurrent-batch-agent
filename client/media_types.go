// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": Application Media Types
//
// Command:
// $ goagen
// --design=github.com/groovenauts/blocks-concurrent-batch-server/design
// --out=$(GOPATH)/src/github.com/groovenauts/blocks-concurrent-batch-server
// --version=v1.3.1

package client

import (
	"github.com/goadesign/goa"
	"net/http"
	"time"
)

// Dummy auth (default view)
//
// Identifier: application/vnd.dummy-auth+json; view=default
type DummyAuth struct {
	OrganizationID string `form:"organization_id" json:"organization_id" yaml:"organization_id" xml:"organization_id"`
	Token          string `form:"token" json:"token" yaml:"token" xml:"token"`
}

// Validate validates the DummyAuth media type instance.
func (mt *DummyAuth) Validate() (err error) {
	if mt.OrganizationID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "organization_id"))
	}
	if mt.Token == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "token"))
	}
	return
}

// DecodeDummyAuth decodes the DummyAuth instance encoded in resp body.
func (c *Client) DecodeDummyAuth(resp *http.Response) (*DummyAuth, error) {
	var decoded DummyAuth
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// DecodeErrorResponse decodes the ErrorResponse instance encoded in resp body.
func (c *Client) DecodeErrorResponse(resp *http.Response) (*goa.ErrorResponse, error) {
	var decoded goa.ErrorResponse
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// instance-group (default view)
//
// Identifier: application/vnd.instance-group+json; view=default
type InstanceGroup struct {
	// Boot disk
	BootDisk *InstanceGroupVMDisk `form:"boot_disk" json:"boot_disk" yaml:"boot_disk" xml:"boot_disk"`
	// Datetime created
	CreatedAt *time.Time `form:"created_at,omitempty" json:"created_at,omitempty" yaml:"created_at,omitempty" xml:"created_at,omitempty"`
	// Deployment name
	DeploymentName string `form:"deployment_name" json:"deployment_name" yaml:"deployment_name" xml:"deployment_name"`
	// GPU Accelerators
	GpuAccelerators *InstanceGroupAccelerators `form:"gpu_accelerators" json:"gpu_accelerators" yaml:"gpu_accelerators" xml:"gpu_accelerators"`
	// ID
	ID string `form:"id" json:"id" yaml:"id" xml:"id"`
	// Instance size
	InstanceSize int `form:"instance_size" json:"instance_size" yaml:"instance_size" xml:"instance_size"`
	// GCE Machine Type
	MachineType string `form:"machine_type" json:"machine_type" yaml:"machine_type" xml:"machine_type"`
	// Name
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Use preemptible VMs
	Preemptible bool `form:"preemptible" json:"preemptible" yaml:"preemptible" xml:"preemptible"`
	// GCP Project ID
	ProjectID string `form:"project_id" json:"project_id" yaml:"project_id" xml:"project_id"`
	// Startup script
	StartupScript string `form:"startup_script" json:"startup_script" yaml:"startup_script" xml:"startup_script"`
	// Instance Group Status
	Status string `form:"status" json:"status" yaml:"status" xml:"status"`
	// Token Consumption
	TokenConsumption int `form:"token_consumption" json:"token_consumption" yaml:"token_consumption" xml:"token_consumption"`
	// Datetime updated
	UpdatedAt *time.Time `form:"updated_at,omitempty" json:"updated_at,omitempty" yaml:"updated_at,omitempty" xml:"updated_at,omitempty"`
	// GCP zone
	Zone string `form:"zone" json:"zone" yaml:"zone" xml:"zone"`
}

// Validate validates the InstanceGroup media type instance.
func (mt *InstanceGroup) Validate() (err error) {
	if mt.ID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id"))
	}
	if mt.Status == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "status"))
	}
	if mt.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if mt.ProjectID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "project_id"))
	}
	if mt.Zone == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "zone"))
	}
	if mt.BootDisk == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "boot_disk"))
	}
	if mt.MachineType == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "machine_type"))
	}
	if mt.GpuAccelerators == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "gpu_accelerators"))
	}

	if mt.StartupScript == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "startup_script"))
	}
	if mt.DeploymentName == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "deployment_name"))
	}

	if mt.BootDisk != nil {
		if err2 := mt.BootDisk.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if mt.GpuAccelerators != nil {
		if err2 := mt.GpuAccelerators.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if !(mt.Status == "constructing" || mt.Status == "constructing_error" || mt.Status == "constructed" || mt.Status == "resizing" || mt.Status == "destructing" || mt.Status == "destructing_error" || mt.Status == "destructed") {
		err = goa.MergeErrors(err, goa.InvalidEnumValueError(`response.status`, mt.Status, []interface{}{"constructing", "constructing_error", "constructed", "resizing", "destructing", "destructing_error", "destructed"}))
	}
	return
}

// DecodeInstanceGroup decodes the InstanceGroup instance encoded in resp body.
func (c *Client) DecodeInstanceGroup(resp *http.Response) (*InstanceGroup, error) {
	var decoded InstanceGroup
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// instance-group-operation (default view)
//
// Identifier: application/vnd.instance-group-operation+json; view=default
type InstanceGroupOperation struct {
	// Datetime created
	CreatedAt *time.Time `form:"created_at,omitempty" json:"created_at,omitempty" yaml:"created_at,omitempty" xml:"created_at,omitempty"`
	// ID
	ID string `form:"id" json:"id" yaml:"id" xml:"id"`
	// Name
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Operation Type
	OperationType string `form:"operation_type" json:"operation_type" yaml:"operation_type" xml:"operation_type"`
	// Owner id
	OwnerID string `form:"owner_id" json:"owner_id" yaml:"owner_id" xml:"owner_id"`
	// Owner type name
	OwnerType string `form:"owner_type" json:"owner_type" yaml:"owner_type" xml:"owner_type"`
	// GCP Project ID
	ProjectID string `form:"project_id" json:"project_id" yaml:"project_id" xml:"project_id"`
	// Service name
	Service string `form:"service" json:"service" yaml:"service" xml:"service"`
	// Operation Status
	Status string `form:"status" json:"status" yaml:"status" xml:"status"`
	// Datetime updated
	UpdatedAt *time.Time `form:"updated_at,omitempty" json:"updated_at,omitempty" yaml:"updated_at,omitempty" xml:"updated_at,omitempty"`
	// GCP zone
	Zone string `form:"zone" json:"zone" yaml:"zone" xml:"zone"`
}

// Validate validates the InstanceGroupOperation media type instance.
func (mt *InstanceGroupOperation) Validate() (err error) {
	if mt.ID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id"))
	}
	if mt.OwnerType == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "owner_type"))
	}
	if mt.OwnerID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "owner_id"))
	}
	if mt.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if mt.Service == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "service"))
	}
	if mt.OperationType == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "operation_type"))
	}
	if mt.Status == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "status"))
	}
	if mt.ProjectID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "project_id"))
	}
	if mt.Zone == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "zone"))
	}
	return
}

// DecodeInstanceGroupOperation decodes the InstanceGroupOperation instance encoded in resp body.
func (c *Client) DecodeInstanceGroupOperation(resp *http.Response) (*InstanceGroupOperation, error) {
	var decoded InstanceGroupOperation
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// Instance-GroupCollection is the media type for an array of Instance-Group (default view)
//
// Identifier: application/vnd.instance-group+json; type=collection; view=default
type InstanceGroupCollection []*InstanceGroup

// Validate validates the InstanceGroupCollection media type instance.
func (mt InstanceGroupCollection) Validate() (err error) {
	for _, e := range mt {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// DecodeInstanceGroupCollection decodes the InstanceGroupCollection instance encoded in resp body.
func (c *Client) DecodeInstanceGroupCollection(resp *http.Response) (InstanceGroupCollection, error) {
	var decoded InstanceGroupCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// job (default view)
//
// Identifier: application/vnd.job+json; view=default
type Job struct {
	// Datetime created
	CreatedAt *time.Time `form:"created_at,omitempty" json:"created_at,omitempty" yaml:"created_at,omitempty" xml:"created_at,omitempty"`
	// Time when job finishes
	FinishTime *time.Time `form:"finish_time,omitempty" json:"finish_time,omitempty" yaml:"finish_time,omitempty" xml:"finish_time,omitempty"`
	// Hostname where job is running
	HostName *string `form:"host_name,omitempty" json:"host_name,omitempty" yaml:"host_name,omitempty" xml:"host_name,omitempty"`
	// ID
	ID string `form:"id" json:"id" yaml:"id" xml:"id"`
	// ID assigned by client
	IDByClient string `form:"id_by_client" json:"id_by_client" yaml:"id_by_client" xml:"id_by_client"`
	// Job message
	Message *JobMessage `form:"message" json:"message" yaml:"message" xml:"message"`
	// Pubsub Message ID
	MessageID *string `form:"message_id,omitempty" json:"message_id,omitempty" yaml:"message_id,omitempty" xml:"message_id,omitempty"`
	// PipelineBase ID (UUID)
	PipelineBaseID *string `form:"pipeline_base_id,omitempty" json:"pipeline_base_id,omitempty" yaml:"pipeline_base_id,omitempty" xml:"pipeline_base_id,omitempty"`
	// Pipeline ID (UUID)
	PipelineID *string `form:"pipeline_id,omitempty" json:"pipeline_id,omitempty" yaml:"pipeline_id,omitempty" xml:"pipeline_id,omitempty"`
	// Time when job is published
	PublishTime *time.Time `form:"publish_time,omitempty" json:"publish_time,omitempty" yaml:"publish_time,omitempty" xml:"publish_time,omitempty"`
	// Time when job starts
	StartTime *time.Time `form:"start_time,omitempty" json:"start_time,omitempty" yaml:"start_time,omitempty" xml:"start_time,omitempty"`
	// Job Status
	Status *string `form:"status,omitempty" json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	// Datetime updated
	UpdatedAt *time.Time `form:"updated_at,omitempty" json:"updated_at,omitempty" yaml:"updated_at,omitempty" xml:"updated_at,omitempty"`
}

// Validate validates the Job media type instance.
func (mt *Job) Validate() (err error) {
	if mt.ID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id"))
	}
	if mt.Message == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "message"))
	}
	if mt.IDByClient == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id_by_client"))
	}
	if mt.Status != nil {
		if !(*mt.Status == "inactive" || *mt.Status == "blocked" || *mt.Status == "publishing" || *mt.Status == "publishing_error" || *mt.Status == "published" || *mt.Status == "started" || *mt.Status == "success" || *mt.Status == "failure") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError(`response.status`, *mt.Status, []interface{}{"inactive", "blocked", "publishing", "publishing_error", "published", "started", "success", "failure"}))
		}
	}
	return
}

// DecodeJob decodes the Job instance encoded in resp body.
func (c *Client) DecodeJob(resp *http.Response) (*Job, error) {
	var decoded Job
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// pipeline (default view)
//
// Identifier: application/vnd.pipeline+json; view=default
type Pipeline struct {
	// Container configuration
	Container *PipelineContainerPayload `form:"container" json:"container" yaml:"container" xml:"container"`
	// Datetime created
	CreatedAt *time.Time `form:"created_at,omitempty" json:"created_at,omitempty" yaml:"created_at,omitempty" xml:"created_at,omitempty"`
	// Current pipeline base ID
	CurrBaseID *string `form:"curr_base_id,omitempty" json:"curr_base_id,omitempty" yaml:"curr_base_id,omitempty" xml:"curr_base_id,omitempty"`
	// Hibernation delay in seconds since last job finsihed
	HibernationDelay int `form:"hibernation_delay" json:"hibernation_delay" yaml:"hibernation_delay" xml:"hibernation_delay"`
	// ID
	ID *string `form:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty" xml:"id,omitempty"`
	// Instance Group configuration
	InstanceGroup *InstanceGroupPayloadBody `form:"instance_group" json:"instance_group" yaml:"instance_group" xml:"instance_group"`
	// Name of pipeline_base
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Next pipeline base ID
	NextBaseID *string `form:"next_base_id,omitempty" json:"next_base_id,omitempty" yaml:"next_base_id,omitempty" xml:"next_base_id,omitempty"`
	// Previous pipeline base ID
	PrevBaseID *string `form:"prev_base_id,omitempty" json:"prev_base_id,omitempty" yaml:"prev_base_id,omitempty" xml:"prev_base_id,omitempty"`
	// Pipeline Status
	Status *string `form:"status,omitempty" json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	// Datetime updated
	UpdatedAt *time.Time `form:"updated_at,omitempty" json:"updated_at,omitempty" yaml:"updated_at,omitempty" xml:"updated_at,omitempty"`
}

// Validate validates the Pipeline media type instance.
func (mt *Pipeline) Validate() (err error) {
	if mt.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if mt.InstanceGroup == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "instance_group"))
	}
	if mt.Container == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "container"))
	}

	if mt.Container != nil {
		if err2 := mt.Container.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if mt.InstanceGroup != nil {
		if err2 := mt.InstanceGroup.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if mt.Status != nil {
		if !(*mt.Status == "current_preparing" || *mt.Status == "current_preparing_error" || *mt.Status == "running" || *mt.Status == "next_preparing" || *mt.Status == "stopping" || *mt.Status == "stopping_error" || *mt.Status == "stopped") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError(`response.status`, *mt.Status, []interface{}{"current_preparing", "current_preparing_error", "running", "next_preparing", "stopping", "stopping_error", "stopped"}))
		}
	}
	return
}

// DecodePipeline decodes the Pipeline instance encoded in resp body.
func (c *Client) DecodePipeline(resp *http.Response) (*Pipeline, error) {
	var decoded Pipeline
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// pipeline-base (default view)
//
// Identifier: application/vnd.pipeline-base+json; view=default
type PipelineBase struct {
	// Container configuration
	Container *PipelineContainerPayload `form:"container" json:"container" yaml:"container" xml:"container"`
	// Datetime created
	CreatedAt *time.Time `form:"created_at,omitempty" json:"created_at,omitempty" yaml:"created_at,omitempty" xml:"created_at,omitempty"`
	// Hibernation delay in seconds since last job finsihed
	HibernationDelay int `form:"hibernation_delay" json:"hibernation_delay" yaml:"hibernation_delay" xml:"hibernation_delay"`
	// ID
	ID string `form:"id" json:"id" yaml:"id" xml:"id"`
	// Instance Group configuration
	InstanceGroup *InstanceGroupPayloadBody `form:"instance_group" json:"instance_group" yaml:"instance_group" xml:"instance_group"`
	// ID of instance group
	InstanceGroupID *string `form:"instance_group_id,omitempty" json:"instance_group_id,omitempty" yaml:"instance_group_id,omitempty" xml:"instance_group_id,omitempty"`
	// Pipeline Base Status
	Status *string `form:"status,omitempty" json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	// Datetime updated
	UpdatedAt *time.Time `form:"updated_at,omitempty" json:"updated_at,omitempty" yaml:"updated_at,omitempty" xml:"updated_at,omitempty"`
}

// Validate validates the PipelineBase media type instance.
func (mt *PipelineBase) Validate() (err error) {
	if mt.ID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id"))
	}
	if mt.InstanceGroup == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "instance_group"))
	}
	if mt.Container == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "container"))
	}

	if mt.Container != nil {
		if err2 := mt.Container.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if mt.InstanceGroup != nil {
		if err2 := mt.InstanceGroup.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if mt.Status != nil {
		if !(*mt.Status == "opening" || *mt.Status == "opening_error" || *mt.Status == "hibernating" || *mt.Status == "waking" || *mt.Status == "waking_error" || *mt.Status == "awake" || *mt.Status == "hibernation_checking" || *mt.Status == "hibernation_going" || *mt.Status == "hibernation_going_error" || *mt.Status == "closing" || *mt.Status == "closing_error" || *mt.Status == "closed") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError(`response.status`, *mt.Status, []interface{}{"opening", "opening_error", "hibernating", "waking", "waking_error", "awake", "hibernation_checking", "hibernation_going", "hibernation_going_error", "closing", "closing_error", "closed"}))
		}
	}
	return
}

// DecodePipelineBase decodes the PipelineBase instance encoded in resp body.
func (c *Client) DecodePipelineBase(resp *http.Response) (*PipelineBase, error) {
	var decoded PipelineBase
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// Pipeline-BaseCollection is the media type for an array of Pipeline-Base (default view)
//
// Identifier: application/vnd.pipeline-base+json; type=collection; view=default
type PipelineBaseCollection []*PipelineBase

// Validate validates the PipelineBaseCollection media type instance.
func (mt PipelineBaseCollection) Validate() (err error) {
	for _, e := range mt {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// DecodePipelineBaseCollection decodes the PipelineBaseCollection instance encoded in resp body.
func (c *Client) DecodePipelineBaseCollection(resp *http.Response) (PipelineBaseCollection, error) {
	var decoded PipelineBaseCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// PipelineCollection is the media type for an array of Pipeline (default view)
//
// Identifier: application/vnd.pipeline+json; type=collection; view=default
type PipelineCollection []*Pipeline

// Validate validates the PipelineCollection media type instance.
func (mt PipelineCollection) Validate() (err error) {
	for _, e := range mt {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// DecodePipelineCollection decodes the PipelineCollection instance encoded in resp body.
func (c *Client) DecodePipelineCollection(resp *http.Response) (PipelineCollection, error) {
	var decoded PipelineCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}
