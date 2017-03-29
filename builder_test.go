package pipeline

import (
	// "fmt"
	"regexp"
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

	pl := Pipeline{
		Name:          "pipeline01",
		ProjectID:     "dummy-proj-999",
		Zone:          "us-central1-f",
		SourceImage:   "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
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
}

func TestBuildDeployment(t *testing.T) {
	b := &Builder{}
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	pl := Pipeline{
		Name:          "pipeline01",
		ProjectID:     "dummy-proj-999",
		Zone:          "us-central1-f",
		SourceImage:   "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
	}
	d, err := b.BuildDeployment(&pl)
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

func TestBuildStartupScript(t *testing.T) {
	b := &Builder{}
	pl := Pipeline{
		Name:          "pipeline01",
		ProjectID:     "dummy-proj-999",
		Zone:          "us-central1-f",
		SourceImage:   "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "",
	}
	ss := b.buildStartupScript(&pl)
	expected :=
		StartupScriptHeader + "\n" +
			"TIMEOUT=600 with_backoff docker pull groovenauts/batch_type_iot_example:0.3.1\n" +
			"for i in {1..2}; do docker run -d" +
			" -e PROJECT=" + pl.ProjectID +
			" -e PIPELINE=" + pl.Name +
			" -e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref." + pl.Name + "-job-subscription.name)" +
			" -e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref." + pl.Name + "-progress-topic.name)" +
			" " + pl.ContainerName +
			" " + pl.Command +
			" ; done"
	assert.Equal(t, expected, ss)

	// Use cos-cloud project's image
	pl.SourceImage = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-56-9000-84-2"
	ss = b.buildStartupScript(&pl)
	assert.Equal(t, expected, ss)

	// Use cos-cloud project's image and private image in asia.gcr.io
	pl.SourceImage = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-56-9000-84-2"
	pl.ContainerName = "asia.gcr.io/example/test_worker:0.0.1"
	ss = b.buildStartupScript(&pl)
	expected =
		StartupScriptHeader + "\n" +
			"METADATA=http://metadata.google.internal/computeMetadata/v1\n" +
			"SVC_ACCT=$METADATA/instance/service-accounts/default\n" +
			"ACCESS_TOKEN=$(curl -H 'Metadata-Flavor: Google' $SVC_ACCT/token | cut -d'\"' -f 4)\n" +
			"TIMEOUT=60 with_backoff docker --config /home/chronos/.docker login -e 1234@5678.com -u _token -p $ACCESS_TOKEN https://asia.gcr.io\n" +
			"TIMEOUT=600 with_backoff docker --config /home/chronos/.docker pull " + pl.ContainerName + "\n" +
			"for i in {1..2}; do docker --config /home/chronos/.docker run -d" +
			" -e PROJECT=" + pl.ProjectID +
			" -e PIPELINE=" + pl.Name +
			" -e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref." + pl.Name + "-job-subscription.name)" +
			" -e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref." + pl.Name + "-progress-topic.name)" +
			" " + pl.ContainerName +
			" " + pl.Command +
			" ; done"
	//fmt.Println(ss)
	assert.Equal(t, expected, ss)

	// Use cos-cloud project's image and private image in gcr.io
	pl.SourceImage = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-56-9000-84-2"
	pl.ContainerName = "gcr.io/example/test_worker:0.0.1" // NOT from asia.gcr.io
	ss = b.buildStartupScript(&pl)
	re := regexp.MustCompile(`asia.gcr.io`)
	expected = re.ReplaceAllString(expected, "gcr.io")
	//fmt.Println(expected)
	assert.Equal(t, expected, ss)
}
