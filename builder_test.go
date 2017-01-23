package pipeline

import (
	"testing"
	// "golang.org/x/net/context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"encoding/json"
	// "google.golang.org/api/deploymentmanager/v2"
)

func TestGenerateContent(t *testing.T) {
	b := &Builder{}
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	// If ackDeadlineSeconds or targetSize is float64, cast it to int
	for _, r := range expected.Resources {
		for k,v := range r.Properties {
			switch vi := v.(type){
			case float64:
				r.Properties[k] = int(vi)
			}
		}
	}

	plp := PipelineProps{
		Name: "pipeline01",
		ProjectID: "dummy-proj-999",
		Zone: "us-central1-f",
		SourceImage: "https://www.googleapis.com/compute/v1/projects/coreos-cloud/global/images/coreos-stable-1235-6-0-v20170111",
		MachineType: "f1-micro",
		TargetSize: 2,
	}
	result := b.GenerateDeploymentResources(&plp)
	assert.Equal(t, &expected, result)
}

func TestBuildDeployment(t *testing.T) {
	b := &Builder{}
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	plp := PipelineProps{
		Name: "pipeline01",
		ProjectID: "dummy-proj-999",
		Zone: "us-central1-f",
		SourceImage: "https://www.googleapis.com/compute/v1/projects/coreos-cloud/global/images/coreos-stable-1235-6-0-v20170111",
		MachineType: "f1-micro",
		TargetSize: 2,
	}
	d, err := b.BuildDeployment(&plp)
	assert.NoError(t, err)
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L348-L434
	assert.Empty(t, d.Description)
	assert.Empty(t, d.Fingerprint)
	assert.Empty(t, d.Id)
	assert.Empty(t, d.InsertTime)
	assert.Empty(t, d.Labels)
	assert.Empty(t, d.Manifest)
	assert.Equal(t, plp.Name, d.Name)
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
