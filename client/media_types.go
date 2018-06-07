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
)

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
	BootDisk *PipelineVMDisk `form:"boot_disk" json:"boot_disk" yaml:"boot_disk" xml:"boot_disk"`
	// Deployment name
	DeploymentName string `form:"deployment_name" json:"deployment_name" yaml:"deployment_name" xml:"deployment_name"`
	// GPU Accelerators
	GpuAccelerators *Accelerators `form:"gpu_accelerators" json:"gpu_accelerators" yaml:"gpu_accelerators" xml:"gpu_accelerators"`
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
	// Status
	Status string `form:"status" json:"status" yaml:"status" xml:"status"`
	// Token Consumption
	TokenConsumption int `form:"token_consumption" json:"token_consumption" yaml:"token_consumption" xml:"token_consumption"`
	// GCP zone
	Zone string `form:"zone" json:"zone" yaml:"zone" xml:"zone"`
}

// Validate validates the InstanceGroup media type instance.
func (mt *InstanceGroup) Validate() (err error) {
	if mt.ID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "id"))
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
	if mt.Status == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "status"))
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
