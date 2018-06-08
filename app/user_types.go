// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": Application User Types
//
// Command:
// $ goagen
// --design=github.com/groovenauts/blocks-concurrent-batch-server/design
// --out=$(GOPATH)/src/github.com/groovenauts/blocks-concurrent-batch-server
// --version=v1.3.1

package app

import (
	"github.com/goadesign/goa"
)

// accelerators user type.
type accelerators struct {
	// Count
	Count *int `form:"count,omitempty" json:"count,omitempty" yaml:"count,omitempty" xml:"count,omitempty"`
	// Type
	Type *string `form:"type,omitempty" json:"type,omitempty" yaml:"type,omitempty" xml:"type,omitempty"`
}

// Finalize sets the default values for accelerators type instance.
func (ut *accelerators) Finalize() {
	var defaultCount = 0
	if ut.Count == nil {
		ut.Count = &defaultCount
	}
	var defaultType = ""
	if ut.Type == nil {
		ut.Type = &defaultType
	}
}

// Validate validates the accelerators type instance.
func (ut *accelerators) Validate() (err error) {
	if ut.Count == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "count"))
	}
	if ut.Type == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "type"))
	}
	return
}

// Publicize creates Accelerators from accelerators
func (ut *accelerators) Publicize() *Accelerators {
	var pub Accelerators
	if ut.Count != nil {
		pub.Count = *ut.Count
	}
	if ut.Type != nil {
		pub.Type = *ut.Type
	}
	return &pub
}

// Accelerators user type.
type Accelerators struct {
	// Count
	Count int `form:"count" json:"count" yaml:"count" xml:"count"`
	// Type
	Type string `form:"type" json:"type" yaml:"type" xml:"type"`
}

// Validate validates the Accelerators type instance.
func (ut *Accelerators) Validate() (err error) {

	if ut.Type == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "type"))
	}
	return
}

// instanceGroupPayload user type.
type instanceGroupPayload struct {
	// Boot disk
	BootDisk *pipelineVMDisk `form:"boot_disk,omitempty" json:"boot_disk,omitempty" yaml:"boot_disk,omitempty" xml:"boot_disk,omitempty"`
	// Deployment name
	DeploymentName *string `form:"deployment_name,omitempty" json:"deployment_name,omitempty" yaml:"deployment_name,omitempty" xml:"deployment_name,omitempty"`
	// GPU Accelerators
	GpuAccelerators *accelerators `form:"gpu_accelerators,omitempty" json:"gpu_accelerators,omitempty" yaml:"gpu_accelerators,omitempty" xml:"gpu_accelerators,omitempty"`
	// Instance size
	InstanceSize *int `form:"instance_size,omitempty" json:"instance_size,omitempty" yaml:"instance_size,omitempty" xml:"instance_size,omitempty"`
	// GCE Machine Type
	MachineType *string `form:"machine_type,omitempty" json:"machine_type,omitempty" yaml:"machine_type,omitempty" xml:"machine_type,omitempty"`
	// Name
	Name *string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	// Owner pipeline_base id (UUID)
	PipelineBaseID *string `form:"pipeline_base_id,omitempty" json:"pipeline_base_id,omitempty" yaml:"pipeline_base_id,omitempty" xml:"pipeline_base_id,omitempty"`
	// Use preemptible VMs
	Preemptible *bool `form:"preemptible,omitempty" json:"preemptible,omitempty" yaml:"preemptible,omitempty" xml:"preemptible,omitempty"`
	// GCP Project ID
	ProjectID *string `form:"project_id,omitempty" json:"project_id,omitempty" yaml:"project_id,omitempty" xml:"project_id,omitempty"`
	// Startup script
	StartupScript *string `form:"startup_script,omitempty" json:"startup_script,omitempty" yaml:"startup_script,omitempty" xml:"startup_script,omitempty"`
	// Token Consumption
	TokenConsumption *int `form:"token_consumption,omitempty" json:"token_consumption,omitempty" yaml:"token_consumption,omitempty" xml:"token_consumption,omitempty"`
	// GCP zone
	Zone *string `form:"zone,omitempty" json:"zone,omitempty" yaml:"zone,omitempty" xml:"zone,omitempty"`
}

// Finalize sets the default values for instanceGroupPayload type instance.
func (ut *instanceGroupPayload) Finalize() {
	if ut.BootDisk != nil {
		var defaultDiskSizeGb = 0
		if ut.BootDisk.DiskSizeGb == nil {
			ut.BootDisk.DiskSizeGb = &defaultDiskSizeGb
		}
		var defaultDiskType = ""
		if ut.BootDisk.DiskType == nil {
			ut.BootDisk.DiskType = &defaultDiskType
		}
		var defaultSourceImage = ""
		if ut.BootDisk.SourceImage == nil {
			ut.BootDisk.SourceImage = &defaultSourceImage
		}
	}
	if ut.GpuAccelerators != nil {
		var defaultCount = 0
		if ut.GpuAccelerators.Count == nil {
			ut.GpuAccelerators.Count = &defaultCount
		}
		var defaultType = ""
		if ut.GpuAccelerators.Type == nil {
			ut.GpuAccelerators.Type = &defaultType
		}
	}
}

// Validate validates the instanceGroupPayload type instance.
func (ut *instanceGroupPayload) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "name"))
	}
	if ut.ProjectID == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "project_id"))
	}
	if ut.Zone == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "zone"))
	}
	if ut.BootDisk == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "boot_disk"))
	}
	if ut.MachineType == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "machine_type"))
	}
	if ut.BootDisk != nil {
		if err2 := ut.BootDisk.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.GpuAccelerators != nil {
		if err2 := ut.GpuAccelerators.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates InstanceGroupPayload from instanceGroupPayload
func (ut *instanceGroupPayload) Publicize() *InstanceGroupPayload {
	var pub InstanceGroupPayload
	if ut.BootDisk != nil {
		pub.BootDisk = ut.BootDisk.Publicize()
	}
	if ut.DeploymentName != nil {
		pub.DeploymentName = ut.DeploymentName
	}
	if ut.GpuAccelerators != nil {
		pub.GpuAccelerators = ut.GpuAccelerators.Publicize()
	}
	if ut.InstanceSize != nil {
		pub.InstanceSize = ut.InstanceSize
	}
	if ut.MachineType != nil {
		pub.MachineType = *ut.MachineType
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	if ut.PipelineBaseID != nil {
		pub.PipelineBaseID = ut.PipelineBaseID
	}
	if ut.Preemptible != nil {
		pub.Preemptible = ut.Preemptible
	}
	if ut.ProjectID != nil {
		pub.ProjectID = *ut.ProjectID
	}
	if ut.StartupScript != nil {
		pub.StartupScript = ut.StartupScript
	}
	if ut.TokenConsumption != nil {
		pub.TokenConsumption = ut.TokenConsumption
	}
	if ut.Zone != nil {
		pub.Zone = *ut.Zone
	}
	return &pub
}

// InstanceGroupPayload user type.
type InstanceGroupPayload struct {
	// Boot disk
	BootDisk *PipelineVMDisk `form:"boot_disk" json:"boot_disk" yaml:"boot_disk" xml:"boot_disk"`
	// Deployment name
	DeploymentName *string `form:"deployment_name,omitempty" json:"deployment_name,omitempty" yaml:"deployment_name,omitempty" xml:"deployment_name,omitempty"`
	// GPU Accelerators
	GpuAccelerators *Accelerators `form:"gpu_accelerators,omitempty" json:"gpu_accelerators,omitempty" yaml:"gpu_accelerators,omitempty" xml:"gpu_accelerators,omitempty"`
	// Instance size
	InstanceSize *int `form:"instance_size,omitempty" json:"instance_size,omitempty" yaml:"instance_size,omitempty" xml:"instance_size,omitempty"`
	// GCE Machine Type
	MachineType string `form:"machine_type" json:"machine_type" yaml:"machine_type" xml:"machine_type"`
	// Name
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Owner pipeline_base id (UUID)
	PipelineBaseID *string `form:"pipeline_base_id,omitempty" json:"pipeline_base_id,omitempty" yaml:"pipeline_base_id,omitempty" xml:"pipeline_base_id,omitempty"`
	// Use preemptible VMs
	Preemptible *bool `form:"preemptible,omitempty" json:"preemptible,omitempty" yaml:"preemptible,omitempty" xml:"preemptible,omitempty"`
	// GCP Project ID
	ProjectID string `form:"project_id" json:"project_id" yaml:"project_id" xml:"project_id"`
	// Startup script
	StartupScript *string `form:"startup_script,omitempty" json:"startup_script,omitempty" yaml:"startup_script,omitempty" xml:"startup_script,omitempty"`
	// Token Consumption
	TokenConsumption *int `form:"token_consumption,omitempty" json:"token_consumption,omitempty" yaml:"token_consumption,omitempty" xml:"token_consumption,omitempty"`
	// GCP zone
	Zone string `form:"zone" json:"zone" yaml:"zone" xml:"zone"`
}

// Validate validates the InstanceGroupPayload type instance.
func (ut *InstanceGroupPayload) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "name"))
	}
	if ut.ProjectID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "project_id"))
	}
	if ut.Zone == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "zone"))
	}
	if ut.BootDisk == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "boot_disk"))
	}
	if ut.MachineType == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "machine_type"))
	}
	if ut.BootDisk != nil {
		if err2 := ut.BootDisk.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.GpuAccelerators != nil {
		if err2 := ut.GpuAccelerators.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// instanceGroupPayloadBody user type.
type instanceGroupPayloadBody struct {
	// Boot disk
	BootDisk *pipelineVMDisk `form:"boot_disk,omitempty" json:"boot_disk,omitempty" yaml:"boot_disk,omitempty" xml:"boot_disk,omitempty"`
	// Deployment name
	DeploymentName *string `form:"deployment_name,omitempty" json:"deployment_name,omitempty" yaml:"deployment_name,omitempty" xml:"deployment_name,omitempty"`
	// GPU Accelerators
	GpuAccelerators *accelerators `form:"gpu_accelerators,omitempty" json:"gpu_accelerators,omitempty" yaml:"gpu_accelerators,omitempty" xml:"gpu_accelerators,omitempty"`
	// Instance size
	InstanceSize *int `form:"instance_size,omitempty" json:"instance_size,omitempty" yaml:"instance_size,omitempty" xml:"instance_size,omitempty"`
	// GCE Machine Type
	MachineType *string `form:"machine_type,omitempty" json:"machine_type,omitempty" yaml:"machine_type,omitempty" xml:"machine_type,omitempty"`
	// Use preemptible VMs
	Preemptible *bool `form:"preemptible,omitempty" json:"preemptible,omitempty" yaml:"preemptible,omitempty" xml:"preemptible,omitempty"`
	// GCP Project ID
	ProjectID *string `form:"project_id,omitempty" json:"project_id,omitempty" yaml:"project_id,omitempty" xml:"project_id,omitempty"`
	// Startup script
	StartupScript *string `form:"startup_script,omitempty" json:"startup_script,omitempty" yaml:"startup_script,omitempty" xml:"startup_script,omitempty"`
	// Token Consumption
	TokenConsumption *int `form:"token_consumption,omitempty" json:"token_consumption,omitempty" yaml:"token_consumption,omitempty" xml:"token_consumption,omitempty"`
	// GCP zone
	Zone *string `form:"zone,omitempty" json:"zone,omitempty" yaml:"zone,omitempty" xml:"zone,omitempty"`
}

// Finalize sets the default values for instanceGroupPayloadBody type instance.
func (ut *instanceGroupPayloadBody) Finalize() {
	if ut.BootDisk != nil {
		var defaultDiskSizeGb = 0
		if ut.BootDisk.DiskSizeGb == nil {
			ut.BootDisk.DiskSizeGb = &defaultDiskSizeGb
		}
		var defaultDiskType = ""
		if ut.BootDisk.DiskType == nil {
			ut.BootDisk.DiskType = &defaultDiskType
		}
		var defaultSourceImage = ""
		if ut.BootDisk.SourceImage == nil {
			ut.BootDisk.SourceImage = &defaultSourceImage
		}
	}
	if ut.GpuAccelerators != nil {
		var defaultCount = 0
		if ut.GpuAccelerators.Count == nil {
			ut.GpuAccelerators.Count = &defaultCount
		}
		var defaultType = ""
		if ut.GpuAccelerators.Type == nil {
			ut.GpuAccelerators.Type = &defaultType
		}
	}
}

// Validate validates the instanceGroupPayloadBody type instance.
func (ut *instanceGroupPayloadBody) Validate() (err error) {
	if ut.ProjectID == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "project_id"))
	}
	if ut.Zone == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "zone"))
	}
	if ut.BootDisk == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "boot_disk"))
	}
	if ut.MachineType == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "machine_type"))
	}
	if ut.BootDisk != nil {
		if err2 := ut.BootDisk.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.GpuAccelerators != nil {
		if err2 := ut.GpuAccelerators.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates InstanceGroupPayloadBody from instanceGroupPayloadBody
func (ut *instanceGroupPayloadBody) Publicize() *InstanceGroupPayloadBody {
	var pub InstanceGroupPayloadBody
	if ut.BootDisk != nil {
		pub.BootDisk = ut.BootDisk.Publicize()
	}
	if ut.DeploymentName != nil {
		pub.DeploymentName = ut.DeploymentName
	}
	if ut.GpuAccelerators != nil {
		pub.GpuAccelerators = ut.GpuAccelerators.Publicize()
	}
	if ut.InstanceSize != nil {
		pub.InstanceSize = ut.InstanceSize
	}
	if ut.MachineType != nil {
		pub.MachineType = *ut.MachineType
	}
	if ut.Preemptible != nil {
		pub.Preemptible = ut.Preemptible
	}
	if ut.ProjectID != nil {
		pub.ProjectID = *ut.ProjectID
	}
	if ut.StartupScript != nil {
		pub.StartupScript = ut.StartupScript
	}
	if ut.TokenConsumption != nil {
		pub.TokenConsumption = ut.TokenConsumption
	}
	if ut.Zone != nil {
		pub.Zone = *ut.Zone
	}
	return &pub
}

// InstanceGroupPayloadBody user type.
type InstanceGroupPayloadBody struct {
	// Boot disk
	BootDisk *PipelineVMDisk `form:"boot_disk" json:"boot_disk" yaml:"boot_disk" xml:"boot_disk"`
	// Deployment name
	DeploymentName *string `form:"deployment_name,omitempty" json:"deployment_name,omitempty" yaml:"deployment_name,omitempty" xml:"deployment_name,omitempty"`
	// GPU Accelerators
	GpuAccelerators *Accelerators `form:"gpu_accelerators,omitempty" json:"gpu_accelerators,omitempty" yaml:"gpu_accelerators,omitempty" xml:"gpu_accelerators,omitempty"`
	// Instance size
	InstanceSize *int `form:"instance_size,omitempty" json:"instance_size,omitempty" yaml:"instance_size,omitempty" xml:"instance_size,omitempty"`
	// GCE Machine Type
	MachineType string `form:"machine_type" json:"machine_type" yaml:"machine_type" xml:"machine_type"`
	// Use preemptible VMs
	Preemptible *bool `form:"preemptible,omitempty" json:"preemptible,omitempty" yaml:"preemptible,omitempty" xml:"preemptible,omitempty"`
	// GCP Project ID
	ProjectID string `form:"project_id" json:"project_id" yaml:"project_id" xml:"project_id"`
	// Startup script
	StartupScript *string `form:"startup_script,omitempty" json:"startup_script,omitempty" yaml:"startup_script,omitempty" xml:"startup_script,omitempty"`
	// Token Consumption
	TokenConsumption *int `form:"token_consumption,omitempty" json:"token_consumption,omitempty" yaml:"token_consumption,omitempty" xml:"token_consumption,omitempty"`
	// GCP zone
	Zone string `form:"zone" json:"zone" yaml:"zone" xml:"zone"`
}

// Validate validates the InstanceGroupPayloadBody type instance.
func (ut *InstanceGroupPayloadBody) Validate() (err error) {
	if ut.ProjectID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "project_id"))
	}
	if ut.Zone == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "zone"))
	}
	if ut.BootDisk == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "boot_disk"))
	}
	if ut.MachineType == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "machine_type"))
	}
	if ut.BootDisk != nil {
		if err2 := ut.BootDisk.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.GpuAccelerators != nil {
		if err2 := ut.GpuAccelerators.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// jobMessage user type.
type jobMessage struct {
	// Attributes
	Attributes map[string]string `form:"attributes,omitempty" json:"attributes,omitempty" yaml:"attributes,omitempty" xml:"attributes,omitempty"`
	// Data
	Data *string `form:"data,omitempty" json:"data,omitempty" yaml:"data,omitempty" xml:"data,omitempty"`
}

// Publicize creates JobMessage from jobMessage
func (ut *jobMessage) Publicize() *JobMessage {
	var pub JobMessage
	if ut.Attributes != nil {
		pub.Attributes = ut.Attributes
	}
	if ut.Data != nil {
		pub.Data = ut.Data
	}
	return &pub
}

// JobMessage user type.
type JobMessage struct {
	// Attributes
	Attributes map[string]string `form:"attributes,omitempty" json:"attributes,omitempty" yaml:"attributes,omitempty" xml:"attributes,omitempty"`
	// Data
	Data *string `form:"data,omitempty" json:"data,omitempty" yaml:"data,omitempty" xml:"data,omitempty"`
}

// jobPayload user type.
type jobPayload struct {
	// ID assigned by client
	IDByClient *string `form:"id_by_client,omitempty" json:"id_by_client,omitempty" yaml:"id_by_client,omitempty" xml:"id_by_client,omitempty"`
	// Job message
	Message *jobMessage `form:"message,omitempty" json:"message,omitempty" yaml:"message,omitempty" xml:"message,omitempty"`
}

// Validate validates the jobPayload type instance.
func (ut *jobPayload) Validate() (err error) {
	if ut.Message == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "message"))
	}
	return
}

// Publicize creates JobPayload from jobPayload
func (ut *jobPayload) Publicize() *JobPayload {
	var pub JobPayload
	if ut.IDByClient != nil {
		pub.IDByClient = ut.IDByClient
	}
	if ut.Message != nil {
		pub.Message = ut.Message.Publicize()
	}
	return &pub
}

// JobPayload user type.
type JobPayload struct {
	// ID assigned by client
	IDByClient *string `form:"id_by_client,omitempty" json:"id_by_client,omitempty" yaml:"id_by_client,omitempty" xml:"id_by_client,omitempty"`
	// Job message
	Message *JobMessage `form:"message" json:"message" yaml:"message" xml:"message"`
}

// Validate validates the JobPayload type instance.
func (ut *JobPayload) Validate() (err error) {
	if ut.Message == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "message"))
	}
	return
}

// operationPayload user type.
type operationPayload struct {
	// Callback task path
	FinalizeTaskPath *string `form:"finalize_task_path,omitempty" json:"finalize_task_path,omitempty" yaml:"finalize_task_path,omitempty" xml:"finalize_task_path,omitempty"`
	// Name
	Name *string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	// Operation Type
	OperationType *string `form:"operation_type,omitempty" json:"operation_type,omitempty" yaml:"operation_type,omitempty" xml:"operation_type,omitempty"`
	// Owner id
	OwnerID *string `form:"owner_id,omitempty" json:"owner_id,omitempty" yaml:"owner_id,omitempty" xml:"owner_id,omitempty"`
	// Owner type name
	OwnerType *string `form:"owner_type,omitempty" json:"owner_type,omitempty" yaml:"owner_type,omitempty" xml:"owner_type,omitempty"`
	// GCP Project ID
	ProjectID *string `form:"project_id,omitempty" json:"project_id,omitempty" yaml:"project_id,omitempty" xml:"project_id,omitempty"`
	// Service name
	Service *string `form:"service,omitempty" json:"service,omitempty" yaml:"service,omitempty" xml:"service,omitempty"`
	// status
	Status *string `form:"status,omitempty" json:"status,omitempty" yaml:"status,omitempty" xml:"status,omitempty"`
	// GCP zone
	Zone *string `form:"zone,omitempty" json:"zone,omitempty" yaml:"zone,omitempty" xml:"zone,omitempty"`
}

// Validate validates the operationPayload type instance.
func (ut *operationPayload) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "name"))
	}
	if ut.Service == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "service"))
	}
	if ut.OperationType == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "operation_type"))
	}
	if ut.Status == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "status"))
	}
	if ut.ProjectID == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "project_id"))
	}
	if ut.Zone == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "zone"))
	}
	return
}

// Publicize creates OperationPayload from operationPayload
func (ut *operationPayload) Publicize() *OperationPayload {
	var pub OperationPayload
	if ut.FinalizeTaskPath != nil {
		pub.FinalizeTaskPath = ut.FinalizeTaskPath
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	if ut.OperationType != nil {
		pub.OperationType = *ut.OperationType
	}
	if ut.OwnerID != nil {
		pub.OwnerID = ut.OwnerID
	}
	if ut.OwnerType != nil {
		pub.OwnerType = ut.OwnerType
	}
	if ut.ProjectID != nil {
		pub.ProjectID = *ut.ProjectID
	}
	if ut.Service != nil {
		pub.Service = *ut.Service
	}
	if ut.Status != nil {
		pub.Status = *ut.Status
	}
	if ut.Zone != nil {
		pub.Zone = *ut.Zone
	}
	return &pub
}

// OperationPayload user type.
type OperationPayload struct {
	// Callback task path
	FinalizeTaskPath *string `form:"finalize_task_path,omitempty" json:"finalize_task_path,omitempty" yaml:"finalize_task_path,omitempty" xml:"finalize_task_path,omitempty"`
	// Name
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Operation Type
	OperationType string `form:"operation_type" json:"operation_type" yaml:"operation_type" xml:"operation_type"`
	// Owner id
	OwnerID *string `form:"owner_id,omitempty" json:"owner_id,omitempty" yaml:"owner_id,omitempty" xml:"owner_id,omitempty"`
	// Owner type name
	OwnerType *string `form:"owner_type,omitempty" json:"owner_type,omitempty" yaml:"owner_type,omitempty" xml:"owner_type,omitempty"`
	// GCP Project ID
	ProjectID string `form:"project_id" json:"project_id" yaml:"project_id" xml:"project_id"`
	// Service name
	Service string `form:"service" json:"service" yaml:"service" xml:"service"`
	// status
	Status string `form:"status" json:"status" yaml:"status" xml:"status"`
	// GCP zone
	Zone string `form:"zone" json:"zone" yaml:"zone" xml:"zone"`
}

// Validate validates the OperationPayload type instance.
func (ut *OperationPayload) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "name"))
	}
	if ut.Service == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "service"))
	}
	if ut.OperationType == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "operation_type"))
	}
	if ut.Status == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "status"))
	}
	if ut.ProjectID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "project_id"))
	}
	if ut.Zone == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "zone"))
	}
	return
}

// pipelineBasePayload user type.
type pipelineBasePayload struct {
	// Container configuration
	Container *pipelineContainerPayload `form:"container,omitempty" json:"container,omitempty" yaml:"container,omitempty" xml:"container,omitempty"`
	// Hibernation delay in seconds since last job finsihed
	HibernationDelay *int `form:"hibernation_delay,omitempty" json:"hibernation_delay,omitempty" yaml:"hibernation_delay,omitempty" xml:"hibernation_delay,omitempty"`
	// Instance Group configuration
	InstanceGroup *instanceGroupPayloadBody `form:"instance_group,omitempty" json:"instance_group,omitempty" yaml:"instance_group,omitempty" xml:"instance_group,omitempty"`
	// Name of pipeline_base
	Name *string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
}

// Finalize sets the default values for pipelineBasePayload type instance.
func (ut *pipelineBasePayload) Finalize() {
	if ut.Container != nil {
		var defaultSize = 1
		if ut.Container.Size == nil {
			ut.Container.Size = &defaultSize
		}
	}
	if ut.InstanceGroup != nil {
		if ut.InstanceGroup.BootDisk != nil {
			var defaultDiskSizeGb = 0
			if ut.InstanceGroup.BootDisk.DiskSizeGb == nil {
				ut.InstanceGroup.BootDisk.DiskSizeGb = &defaultDiskSizeGb
			}
			var defaultDiskType = ""
			if ut.InstanceGroup.BootDisk.DiskType == nil {
				ut.InstanceGroup.BootDisk.DiskType = &defaultDiskType
			}
			var defaultSourceImage = ""
			if ut.InstanceGroup.BootDisk.SourceImage == nil {
				ut.InstanceGroup.BootDisk.SourceImage = &defaultSourceImage
			}
		}
		if ut.InstanceGroup.GpuAccelerators != nil {
			var defaultCount = 0
			if ut.InstanceGroup.GpuAccelerators.Count == nil {
				ut.InstanceGroup.GpuAccelerators.Count = &defaultCount
			}
			var defaultType = ""
			if ut.InstanceGroup.GpuAccelerators.Type == nil {
				ut.InstanceGroup.GpuAccelerators.Type = &defaultType
			}
		}
	}
}

// Validate validates the pipelineBasePayload type instance.
func (ut *pipelineBasePayload) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "name"))
	}
	if ut.InstanceGroup == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "instance_group"))
	}
	if ut.Container == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "container"))
	}
	if ut.Container != nil {
		if err2 := ut.Container.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.InstanceGroup != nil {
		if err2 := ut.InstanceGroup.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates PipelineBasePayload from pipelineBasePayload
func (ut *pipelineBasePayload) Publicize() *PipelineBasePayload {
	var pub PipelineBasePayload
	if ut.Container != nil {
		pub.Container = ut.Container.Publicize()
	}
	if ut.HibernationDelay != nil {
		pub.HibernationDelay = ut.HibernationDelay
	}
	if ut.InstanceGroup != nil {
		pub.InstanceGroup = ut.InstanceGroup.Publicize()
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	return &pub
}

// PipelineBasePayload user type.
type PipelineBasePayload struct {
	// Container configuration
	Container *PipelineContainerPayload `form:"container" json:"container" yaml:"container" xml:"container"`
	// Hibernation delay in seconds since last job finsihed
	HibernationDelay *int `form:"hibernation_delay,omitempty" json:"hibernation_delay,omitempty" yaml:"hibernation_delay,omitempty" xml:"hibernation_delay,omitempty"`
	// Instance Group configuration
	InstanceGroup *InstanceGroupPayloadBody `form:"instance_group" json:"instance_group" yaml:"instance_group" xml:"instance_group"`
	// Name of pipeline_base
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
}

// Validate validates the PipelineBasePayload type instance.
func (ut *PipelineBasePayload) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "name"))
	}
	if ut.InstanceGroup == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "instance_group"))
	}
	if ut.Container == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "container"))
	}
	if ut.Container != nil {
		if err2 := ut.Container.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.InstanceGroup != nil {
		if err2 := ut.InstanceGroup.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// pipelineBasePayloadBody user type.
type pipelineBasePayloadBody struct {
	// Container configuration
	Container *pipelineContainerPayload `form:"container,omitempty" json:"container,omitempty" yaml:"container,omitempty" xml:"container,omitempty"`
	// Hibernation delay in seconds since last job finsihed
	HibernationDelay *int `form:"hibernation_delay,omitempty" json:"hibernation_delay,omitempty" yaml:"hibernation_delay,omitempty" xml:"hibernation_delay,omitempty"`
	// Instance Group configuration
	InstanceGroup *instanceGroupPayloadBody `form:"instance_group,omitempty" json:"instance_group,omitempty" yaml:"instance_group,omitempty" xml:"instance_group,omitempty"`
}

// Finalize sets the default values for pipelineBasePayloadBody type instance.
func (ut *pipelineBasePayloadBody) Finalize() {
	if ut.Container != nil {
		var defaultSize = 1
		if ut.Container.Size == nil {
			ut.Container.Size = &defaultSize
		}
	}
	if ut.InstanceGroup != nil {
		if ut.InstanceGroup.BootDisk != nil {
			var defaultDiskSizeGb = 0
			if ut.InstanceGroup.BootDisk.DiskSizeGb == nil {
				ut.InstanceGroup.BootDisk.DiskSizeGb = &defaultDiskSizeGb
			}
			var defaultDiskType = ""
			if ut.InstanceGroup.BootDisk.DiskType == nil {
				ut.InstanceGroup.BootDisk.DiskType = &defaultDiskType
			}
			var defaultSourceImage = ""
			if ut.InstanceGroup.BootDisk.SourceImage == nil {
				ut.InstanceGroup.BootDisk.SourceImage = &defaultSourceImage
			}
		}
		if ut.InstanceGroup.GpuAccelerators != nil {
			var defaultCount = 0
			if ut.InstanceGroup.GpuAccelerators.Count == nil {
				ut.InstanceGroup.GpuAccelerators.Count = &defaultCount
			}
			var defaultType = ""
			if ut.InstanceGroup.GpuAccelerators.Type == nil {
				ut.InstanceGroup.GpuAccelerators.Type = &defaultType
			}
		}
	}
}

// Validate validates the pipelineBasePayloadBody type instance.
func (ut *pipelineBasePayloadBody) Validate() (err error) {
	if ut.InstanceGroup == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "instance_group"))
	}
	if ut.Container == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "container"))
	}
	if ut.Container != nil {
		if err2 := ut.Container.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.InstanceGroup != nil {
		if err2 := ut.InstanceGroup.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates PipelineBasePayloadBody from pipelineBasePayloadBody
func (ut *pipelineBasePayloadBody) Publicize() *PipelineBasePayloadBody {
	var pub PipelineBasePayloadBody
	if ut.Container != nil {
		pub.Container = ut.Container.Publicize()
	}
	if ut.HibernationDelay != nil {
		pub.HibernationDelay = ut.HibernationDelay
	}
	if ut.InstanceGroup != nil {
		pub.InstanceGroup = ut.InstanceGroup.Publicize()
	}
	return &pub
}

// PipelineBasePayloadBody user type.
type PipelineBasePayloadBody struct {
	// Container configuration
	Container *PipelineContainerPayload `form:"container" json:"container" yaml:"container" xml:"container"`
	// Hibernation delay in seconds since last job finsihed
	HibernationDelay *int `form:"hibernation_delay,omitempty" json:"hibernation_delay,omitempty" yaml:"hibernation_delay,omitempty" xml:"hibernation_delay,omitempty"`
	// Instance Group configuration
	InstanceGroup *InstanceGroupPayloadBody `form:"instance_group" json:"instance_group" yaml:"instance_group" xml:"instance_group"`
}

// Validate validates the PipelineBasePayloadBody type instance.
func (ut *PipelineBasePayloadBody) Validate() (err error) {
	if ut.InstanceGroup == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "instance_group"))
	}
	if ut.Container == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "container"))
	}
	if ut.Container != nil {
		if err2 := ut.Container.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if ut.InstanceGroup != nil {
		if err2 := ut.InstanceGroup.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// pipelineContainerPayload user type.
type pipelineContainerPayload struct {
	// Command for docker run
	Command *string `form:"command,omitempty" json:"command,omitempty" yaml:"command,omitempty" xml:"command,omitempty"`
	// Container name
	Name *string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	// Options for docker run
	Options *string `form:"options,omitempty" json:"options,omitempty" yaml:"options,omitempty" xml:"options,omitempty"`
	// Container size per VM
	Size *int `form:"size,omitempty" json:"size,omitempty" yaml:"size,omitempty" xml:"size,omitempty"`
	// Use stackdriver agent
	StackdriverAgent *bool `form:"stackdriver_agent,omitempty" json:"stackdriver_agent,omitempty" yaml:"stackdriver_agent,omitempty" xml:"stackdriver_agent,omitempty"`
}

// Finalize sets the default values for pipelineContainerPayload type instance.
func (ut *pipelineContainerPayload) Finalize() {
	var defaultSize = 1
	if ut.Size == nil {
		ut.Size = &defaultSize
	}
}

// Validate validates the pipelineContainerPayload type instance.
func (ut *pipelineContainerPayload) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "name"))
	}
	return
}

// Publicize creates PipelineContainerPayload from pipelineContainerPayload
func (ut *pipelineContainerPayload) Publicize() *PipelineContainerPayload {
	var pub PipelineContainerPayload
	if ut.Command != nil {
		pub.Command = ut.Command
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	if ut.Options != nil {
		pub.Options = ut.Options
	}
	if ut.Size != nil {
		pub.Size = *ut.Size
	}
	if ut.StackdriverAgent != nil {
		pub.StackdriverAgent = ut.StackdriverAgent
	}
	return &pub
}

// PipelineContainerPayload user type.
type PipelineContainerPayload struct {
	// Command for docker run
	Command *string `form:"command,omitempty" json:"command,omitempty" yaml:"command,omitempty" xml:"command,omitempty"`
	// Container name
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
	// Options for docker run
	Options *string `form:"options,omitempty" json:"options,omitempty" yaml:"options,omitempty" xml:"options,omitempty"`
	// Container size per VM
	Size int `form:"size" json:"size" yaml:"size" xml:"size"`
	// Use stackdriver agent
	StackdriverAgent *bool `form:"stackdriver_agent,omitempty" json:"stackdriver_agent,omitempty" yaml:"stackdriver_agent,omitempty" xml:"stackdriver_agent,omitempty"`
}

// Validate validates the PipelineContainerPayload type instance.
func (ut *PipelineContainerPayload) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "name"))
	}
	return
}

// pipelinePayload user type.
type pipelinePayload struct {
	// PipelineBase configuration
	Base *pipelineBasePayloadBody `form:"base,omitempty" json:"base,omitempty" yaml:"base,omitempty" xml:"base,omitempty"`
	// Name of pipeline_base
	Name *string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
}

// Finalize sets the default values for pipelinePayload type instance.
func (ut *pipelinePayload) Finalize() {
	if ut.Base != nil {
		if ut.Base.Container != nil {
			var defaultSize = 1
			if ut.Base.Container.Size == nil {
				ut.Base.Container.Size = &defaultSize
			}
		}
		if ut.Base.InstanceGroup != nil {
			if ut.Base.InstanceGroup.BootDisk != nil {
				var defaultDiskSizeGb = 0
				if ut.Base.InstanceGroup.BootDisk.DiskSizeGb == nil {
					ut.Base.InstanceGroup.BootDisk.DiskSizeGb = &defaultDiskSizeGb
				}
				var defaultDiskType = ""
				if ut.Base.InstanceGroup.BootDisk.DiskType == nil {
					ut.Base.InstanceGroup.BootDisk.DiskType = &defaultDiskType
				}
				var defaultSourceImage = ""
				if ut.Base.InstanceGroup.BootDisk.SourceImage == nil {
					ut.Base.InstanceGroup.BootDisk.SourceImage = &defaultSourceImage
				}
			}
			if ut.Base.InstanceGroup.GpuAccelerators != nil {
				var defaultCount = 0
				if ut.Base.InstanceGroup.GpuAccelerators.Count == nil {
					ut.Base.InstanceGroup.GpuAccelerators.Count = &defaultCount
				}
				var defaultType = ""
				if ut.Base.InstanceGroup.GpuAccelerators.Type == nil {
					ut.Base.InstanceGroup.GpuAccelerators.Type = &defaultType
				}
			}
		}
	}
}

// Validate validates the pipelinePayload type instance.
func (ut *pipelinePayload) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "name"))
	}
	if ut.Base == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "base"))
	}
	if ut.Base != nil {
		if err2 := ut.Base.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates PipelinePayload from pipelinePayload
func (ut *pipelinePayload) Publicize() *PipelinePayload {
	var pub PipelinePayload
	if ut.Base != nil {
		pub.Base = ut.Base.Publicize()
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	return &pub
}

// PipelinePayload user type.
type PipelinePayload struct {
	// PipelineBase configuration
	Base *PipelineBasePayloadBody `form:"base" json:"base" yaml:"base" xml:"base"`
	// Name of pipeline_base
	Name string `form:"name" json:"name" yaml:"name" xml:"name"`
}

// Validate validates the PipelinePayload type instance.
func (ut *PipelinePayload) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "name"))
	}
	if ut.Base == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "base"))
	}
	if ut.Base != nil {
		if err2 := ut.Base.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// pipelineVMDisk user type.
type pipelineVMDisk struct {
	// Disk size
	DiskSizeGb *int `form:"disk_size_gb,omitempty" json:"disk_size_gb,omitempty" yaml:"disk_size_gb,omitempty" xml:"disk_size_gb,omitempty"`
	// Disk type
	DiskType *string `form:"disk_type,omitempty" json:"disk_type,omitempty" yaml:"disk_type,omitempty" xml:"disk_type,omitempty"`
	// Source image
	SourceImage *string `form:"source_image,omitempty" json:"source_image,omitempty" yaml:"source_image,omitempty" xml:"source_image,omitempty"`
}

// Finalize sets the default values for pipelineVMDisk type instance.
func (ut *pipelineVMDisk) Finalize() {
	var defaultDiskSizeGb = 0
	if ut.DiskSizeGb == nil {
		ut.DiskSizeGb = &defaultDiskSizeGb
	}
	var defaultDiskType = ""
	if ut.DiskType == nil {
		ut.DiskType = &defaultDiskType
	}
	var defaultSourceImage = ""
	if ut.SourceImage == nil {
		ut.SourceImage = &defaultSourceImage
	}
}

// Validate validates the pipelineVMDisk type instance.
func (ut *pipelineVMDisk) Validate() (err error) {
	if ut.SourceImage == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`request`, "source_image"))
	}
	return
}

// Publicize creates PipelineVMDisk from pipelineVMDisk
func (ut *pipelineVMDisk) Publicize() *PipelineVMDisk {
	var pub PipelineVMDisk
	if ut.DiskSizeGb != nil {
		pub.DiskSizeGb = *ut.DiskSizeGb
	}
	if ut.DiskType != nil {
		pub.DiskType = *ut.DiskType
	}
	if ut.SourceImage != nil {
		pub.SourceImage = *ut.SourceImage
	}
	return &pub
}

// PipelineVMDisk user type.
type PipelineVMDisk struct {
	// Disk size
	DiskSizeGb int `form:"disk_size_gb" json:"disk_size_gb" yaml:"disk_size_gb" xml:"disk_size_gb"`
	// Disk type
	DiskType string `form:"disk_type" json:"disk_type" yaml:"disk_type" xml:"disk_type"`
	// Source image
	SourceImage string `form:"source_image" json:"source_image" yaml:"source_image" xml:"source_image"`
}

// Validate validates the PipelineVMDisk type instance.
func (ut *PipelineVMDisk) Validate() (err error) {
	if ut.SourceImage == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`type`, "source_image"))
	}
	return
}
