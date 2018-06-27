package model

import (
	"fmt"
	"testing"
	// "golang.org/x/net/context"
	"encoding/json"
	"io/ioutil"

	"gopkg.in/go-playground/validator.v9"

	"github.com/stretchr/testify/assert"
	// "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/googleapi"
)

func TestGenerateContent(t *testing.T) {
	b := &Constructor{}
	expected_data, err := ioutil.ReadFile(`constructor_test/pipeline01.json`)
	expected := &Resources{}
	err = json.Unmarshal([]byte(expected_data), expected)
	assert.NoError(t, err)
	// If ackDeadlineSeconds or targetSize is float64, cast it to int
	for _, r := range expected.Resources {
		for k, v := range r.Properties {
			switch vi := v.(type) {
			case float64:
				r.Properties[k] = int(vi)
			}
		}
	}

	pl := InstanceGroup{
		Name:      "pipeline01",
		ProjectID: "dummy-proj-999",
		Zone:      "us-central1-f",
		BootDisk: InstanceGroupVMDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:           "f1-micro",
		InstanceSizeRequested: 2,
	}
	result := b.GenerateDeploymentResources(&pl)

	assert.Equal(t, len(expected.Resources), len(result.Resources))
	for i, expR := range expected.Resources {
		actR := result.Resources[i]
		assert.Equal(t, expR.Type, actR.Type)
		assert.Equal(t, expR.Name, actR.Name)
		for key, expV := range expR.Properties {
			actV := actR.Properties[key]
			assert.Equal(t, expV, actV, "Value for key: "+key)
		}
	}
	assert.Equal(t, expected, result)
	props0 := result.Resources[4].Properties
	assert.IsType(t, map[string]interface{}(nil), props0["properties"])
	props1 := props0["properties"].(map[string]interface{})
	assert.IsType(t, map[string]interface{}(nil), props1["scheduling"])
	props2 := props1["scheduling"].(map[string]interface{})
	assert.Equal(t, false, props2["preemptible"])

	// preemptible
	pl.Preemptible = true
	result = b.GenerateDeploymentResources(&pl)
	props0 = result.Resources[4].Properties
	assert.IsType(t, map[string]interface{}(nil), props0["properties"])
	props1 = props0["properties"].(map[string]interface{})
	assert.IsType(t, map[string]interface{}(nil), props1["scheduling"])
	props2 = props1["scheduling"].(map[string]interface{})
	assert.Equal(t, true, props2["preemptible"])
}

func setupForBuildDeployment() (*Constructor, *InstanceGroup) {
	b := &Constructor{}
	pl := &InstanceGroup{
		Name:      "pipeline01",
		ProjectID: "dummy-proj-999",
		Zone:      "us-central1-f",
		BootDisk: InstanceGroupVMDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:           "f1-micro",
		InstanceSizeRequested: 2,
	}
	return b, pl
}

func TestBuildDeployment(t *testing.T) {
	b, pl := setupForBuildDeployment()
	expected_data, err := ioutil.ReadFile(`constructor_test/pipeline01.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	d, err := b.BuildDeployment(pl)
	assert.NoError(t, err)
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L348-L434
	assert.Empty(t, d.Description)
	assert.Empty(t, d.Fingerprint)
	assert.Empty(t, d.Id)
	assert.Empty(t, d.InsertTime)
	assert.Empty(t, d.Labels)
	assert.Empty(t, d.Manifest)
	assert.Equal(t, pl.Name, d.Name)
	assert.Empty(t, d.Operation)
	assert.Empty(t, d.SelfLink)
	assert.NotEmpty(t, d.Target)
	assert.Empty(t, d.Update)
	assert.Empty(t, d.ForceSendFields)
	assert.Empty(t, d.NullFields)
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L1679-L1709
	tc := d.Target
	assert.NotEmpty(t, tc.Config)
	assert.Empty(t, tc.Imports)
	assert.Empty(t, tc.ForceSendFields)
	assert.Empty(t, tc.NullFields)
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L321-L346
	c := tc.Config
	assert.NotEmpty(t, c.Content)
	assert.Empty(t, c.ForceSendFields)
	assert.Empty(t, c.NullFields)

	actual := Resources{}
	err = json.Unmarshal([]byte(c.Content), &actual)
	assert.NoError(t, err)
	assert.Equal(t, &expected, &actual)
}

func TestBuildDeploymentWithGPU(t *testing.T) {
	b, pl := setupForBuildDeployment()
	pl.GpuAccelerators = InstanceGroupAccelerators{
		Count: 2,
		Type:  "nvidia-tesla-p100",
	}
	pl.Status = ConstructionStarting
	pl.PrepareToCreate()
	err := pl.Validate()
	assert.Error(t, err)
	errors := err.(validator.ValidationErrors)
	fmt.Printf("pl.BootDisk.SourceImage: %v\n", "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/family/ubuntu-1604-lts")
	fmt.Printf("errors: %v\n", errors)
	assert.Equal(t, 1, len(errors))

	// bd := &pl.BootDisk
	// bd.SourceImage = Ubuntu16ImageFamily
	pl.BootDisk.SourceImage = Ubuntu16ImageFamily
	err = pl.Validate()
	assert.NoError(t, err)

	expected_data, err := ioutil.ReadFile(`constructor_test/pipeline02.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	actual := b.buildItProperties(pl)
	assert.NoError(t, err)
	assert.Equal(t, expected.Resources[4].Properties["properties"], actual)
}

func TestBuildBootDisk(t *testing.T) {
	b := &Constructor{}
	d1 := InstanceGroupVMDisk{
		SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
	}
	r1 := b.buildBootDisk(&d1)
	assert.IsType(t, r1["initializeParams"], map[string]interface{}{})
	p1 := r1["initializeParams"].(map[string]interface{})
	assert.Contains(t, p1, "sourceImage")
	assert.NotContains(t, p1, "diskSizeGb")
	assert.NotContains(t, p1, "diskType")

	d2 := InstanceGroupVMDisk{
		DiskSizeGb:  50,
		SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		DiskType:    "projects/dummy-proj-999/zones/asia-east1-a/diskTypes/pd-standard",
	}
	r2 := b.buildBootDisk(&d2)
	assert.IsType(t, r2["initializeParams"], map[string]interface{}{})
	p2 := r2["initializeParams"].(map[string]interface{})
	assert.Contains(t, p2, "sourceImage")
	assert.Contains(t, p2, "diskSizeGb")
	assert.Contains(t, p2, "diskType")
}

func TestGoogleapiError(t *testing.T) {
	// See https://github.com/google/google-api-go-client/blob/master/googleapi/googleapi.go#L114-L135
	msg := "'projects/optical-hangar-158902/global/deployments/pipeline-mjr-59-20170926-163820' already exists and cannot be created., duplicate"
	err := &googleapi.Error{
		Code:    409,
		Message: msg,
	}
	expected := fmt.Sprintf("googleapi: Error %d: %s", err.Code, msg)
	assert.Equal(t, expected, fmt.Sprintf("%v", err))
}
