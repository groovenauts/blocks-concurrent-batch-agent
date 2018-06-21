package models

import (
	"fmt"
	"regexp"
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
		Name:      "pipeline01",
		ProjectID: "dummy-proj-999",
		Zone:      "us-central1-f",
		BootDisk: PipelineVmDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
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

func setupForBuildDeployment() (*Builder, *Pipeline) {
	org1 := &Organization{
		Name:        "org01",
		TokenAmount: 10,
	}

	b := &Builder{}
	pl := &Pipeline{
		Organization: org1,
		Name:         "pipeline01",
		ProjectID:    "dummy-proj-999",
		Zone:         "us-central1-f",
		BootDisk: PipelineVmDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
	}
	return b, pl
}

func TestBuildDeployment(t *testing.T) {
	b, pl := setupForBuildDeployment()
	expected_data, err := ioutil.ReadFile(`builder_test/pipeline01.json`)
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
	pl.GpuAccelerators = Accelerators{
		Count: 2,
		Type:  "nvidia-tesla-p100",
	}
	err := pl.Validate()
	assert.Error(t, err)
	errors := err.(validator.ValidationErrors)
	fmt.Printf("pl.BootDisk.SourceImage: %v\n", pl.BootDisk.SourceImage)
	fmt.Printf("errors: %v\n", errors)
	assert.Equal(t, len(errors), 1)

	// bd := &pl.BootDisk
	// bd.SourceImage = Ubuntu16ImageFamily
	pl.BootDisk.SourceImage = Ubuntu16ImageFamily
	err = pl.Validate()
	assert.NoError(t, err)

	expected_data, err := ioutil.ReadFile(`builder_test/pipeline02.json`)
	expected := Resources{}
	err = json.Unmarshal([]byte(expected_data), &expected)
	assert.NoError(t, err)
	actual := b.buildItProperties(pl)
	assert.NoError(t, err)
	assert.Equal(t, expected.Resources[4].Properties["properties"], actual)
}

func setupTestBuildStartupScript() (*Builder, *Pipeline) {
	b := Builder{}
	pl := Pipeline{
		Name:      "pipeline01",
		ProjectID: "dummy-proj-999",
		Zone:      "us-central1-f",
		BootDisk: PipelineVmDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "",
	}
	return &b, &pl
}

func TestBuildStartupScript1(t *testing.T) {
	b, pl := setupTestBuildStartupScript()
	ss := b.buildStartupScript(pl)
	startupScriptBodyBase :=
		"with_backoff docker pull groovenauts/batch_type_iot_example:0.3.1\n" +
			"for i in {1..2}; do" +
			"\n  docker run -d" +
			" \\\n    -e PROJECT=" + pl.ProjectID +
			" \\\n    -e DOCKER_HOSTNAME=$(hostname)" +
			" \\\n    -e PIPELINE=" + pl.Name +
			" \\\n    -e ZONE=" + pl.Zone +
			" \\\n    -e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref." + pl.Name + "-job-subscription.name)" +
			" \\\n    -e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref." + pl.Name + "-progress-topic.name)"
	startupScriptBody0 := startupScriptBodyBase +
		" \\\n    " + pl.ContainerName +
		" \\\n    " + pl.Command +
		"\ndone"

	assert.Equal(t, StartupScriptHeader+"\n"+startupScriptBody0, ss)

	// Use with DockerRunOptions
	// https://docs.docker.com/engine/reference/commandline/run/#restart-policies---restart
	pl.DockerRunOptions = "--restart=on-failure:3"
	ss1 := b.buildStartupScript(pl)
	startupScriptBody1 := startupScriptBodyBase +
		" \\\n    " + pl.DockerRunOptions +
		" \\\n    " + pl.ContainerName +
		" \\\n    " + pl.Command +
		"\ndone"
	assert.Equal(t, StartupScriptHeader+"\n"+startupScriptBody1, ss1)
	pl.DockerRunOptions = "" // Reset

	// Use cos-cloud project's image
	pl.BootDisk.SourceImage = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-56-9000-84-2"
	ss = b.buildStartupScript(pl)
	assert.Equal(t, StartupScriptHeader+"\n"+startupScriptBody0, ss)

	// Use stackdriver-agent
	pl.StackdriverAgent = true
	assert.Equal(t, StartupScriptHeader+"\n"+StackdriverAgentCommand+"\n"+startupScriptBody0, b.buildStartupScript(pl))
}

func TestBuildStartupScript2(t *testing.T) {
	b, pl := setupTestBuildStartupScript()
	// Use cos-cloud project's image and private image in asia.gcr.io
	pl.BootDisk.SourceImage = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-56-9000-84-2"
	pl.ContainerName = "asia.gcr.io/example/test_worker:0.0.1"
	ss := b.buildStartupScript(pl)
	expected :=
		StartupScriptHeader + "\n" +
			"METADATA=http://metadata.google.internal/computeMetadata/v1" +
			"\nSVC_ACCT=$METADATA/instance/service-accounts/default" +
			"\nACCESS_TOKEN=$(curl -H 'Metadata-Flavor: Google' $SVC_ACCT/token | cut -d'\"' -f 4)" +
			"\nwith_backoff docker --config /home/chronos/.docker login -e 1234@5678.com -u _token -p $ACCESS_TOKEN https://asia.gcr.io" +
			"\nwith_backoff docker --config /home/chronos/.docker pull " + pl.ContainerName +
			"\nfor i in {1..2}; do" +
			"\n  docker --config /home/chronos/.docker run -d" +
			" \\\n    -e PROJECT=" + pl.ProjectID +
			" \\\n    -e DOCKER_HOSTNAME=$(hostname)" +
			" \\\n    -e PIPELINE=" + pl.Name +
			" \\\n    -e ZONE=" + pl.Zone +
			" \\\n    -e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref." + pl.Name + "-job-subscription.name)" +
			" \\\n    -e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref." + pl.Name + "-progress-topic.name)" +
			" \\\n    " + pl.ContainerName +
			" \\\n    " + pl.Command +
			"\ndone"
	//fmt.Println(ss)
	assert.Equal(t, expected, ss)

	// Use cos-cloud project's image and private image in gcr.io
	pl.BootDisk.SourceImage = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-56-9000-84-2"
	pl.ContainerName = "gcr.io/example/test_worker:0.0.1" // NOT from asia.gcr.io
	ss = b.buildStartupScript(pl)
	re := regexp.MustCompile(`asia.gcr.io`)
	expected = re.ReplaceAllString(expected, "gcr.io")
	//fmt.Println(expected)
	assert.Equal(t, expected, ss)
}

func TestBuildBootDisk(t *testing.T) {
	b := &Builder{}
	d1 := PipelineVmDisk{
		SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
	}
	r1 := b.buildBootDisk(&d1)
	assert.IsType(t, r1["initializeParams"], map[string]interface{}{})
	p1 := r1["initializeParams"].(map[string]interface{})
	assert.Contains(t, p1, "sourceImage")
	assert.NotContains(t, p1, "diskSizeGb")
	assert.NotContains(t, p1, "diskType")

	d2 := PipelineVmDisk{
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
