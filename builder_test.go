package pipeline

import (
	"testing"
	// "golang.org/x/net/context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	// "google.golang.org/api/deploymentmanager/v2"
)

func TestGenerateContent(t *testing.T) {
	b := &Builder{}
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	// If ackDeadlineSeconds is float65, cast it to int
	for _, r := range expected.Resources {
		v := r.Properties["ackDeadlineSeconds"]
		switch vi := v.(type) {
		case float64:
			r.Properties["ackDeadlineSeconds"] = int(vi)
		}
	}

	result := b.GenerateDeploymentResources("dummy-proj", "pipeline01")
	assert.Equal(t, &expected, result)
}

func TestBuildDeployment(t *testing.T) {
	b := &Builder{}
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	plp := PipelineProps{
		Name:      "pipeline01",
		ProjectID: "dummy-proj",
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
