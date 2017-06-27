package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"test_utils"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	// "google.golang.org/appengine/log"
)

func TestPipelineJobCRUD(t *testing.T) {
	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if !assert.NoError(t, err) {
		inst.Close()
		return
	}
	ctx := appengine.NewContext(req)

	kinds := []string{"PipelineJobs", "Pipelines", "Organizations"}
	for _, k := range kinds {
		test_utils.ClearDatastore(t, ctx, k)
	}

	org1 := &Organization{Name: "org1"}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	pipelines := map[string]*Pipeline{}
	pipelineNames := []string{"pipeline1", "pipeline2"}
	for _, pipelineName := range pipelineNames {
		pipeline := &Pipeline{
			Organization: org1,
			Name:         pipelineName,
			ProjectID:    "dummy-proj-111",
			Zone:         "asia-northeast1-a",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
			},
			MachineType:   "f1-micro",
			TargetSize:    1,
			ContainerSize: 1,
			ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		}
		err = pipeline.Create(ctx)
		assert.NoError(t, err)
		pipelines[pipelineName] = pipeline

		for i := 1; i < 3; i++ {
			job := &PipelineJob{
				Pipeline:   pipeline,
				IdByClient: fmt.Sprintf("%v-job-%v", pipelineName, i),
				Status:     Published,
				Message: PipelineJobMessage{
					AttributesJson: fmt.Sprintf(`{"foo":"%v"}`, i),
				},
			}
			err = job.Create(ctx)
			assert.NoError(t, err)
		}
	}

	jobs, err := GlobalPipelineJobAccessor.All(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(jobs))

	pipeline1 := pipelines["pipeline1"]
	accessor := pipeline1.JobAccessor()
	jobs, err = accessor.All(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobs))

	// Invalid Job
	invalidJsonPatterns := []string{
		`INVALID JSON DATA`,
		`"VALID JSON String"`,
		`["VALID JSON String Array"]`,
		`{"VALID JSON String to Integer": 1000}`,
	}
	for _, ptn := range invalidJsonPatterns {
		pj := &PipelineJob{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%v-job-Invalid", pipeline1.Name),
			Message: PipelineJobMessage{
				AttributesJson: ptn,
			},
		}
		err = pj.Validate()
		assert.Error(t, err)
	}

	originalPublisher := GlobalPublisher
	dummyPublisher := &DummyPublisher{
		Invocations: []*PublishInvocation{},
	}
	GlobalPublisher = dummyPublisher

	defer func() {
		GlobalPublisher = originalPublisher
	}()

	// CreateAndPublishIfPossible
	attrs := map[string]string{
		"download_files": "gcs://bucket1/path/to/file1",
	}
	attrs_json, err := json.Marshal(attrs)
	assert.NoError(t, err)

	// Don't publish Job Message soon when the pipeline isn't Opened
	for _, st := range []Status{Initialized, Building, Deploying} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &PipelineJob{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-waiting-%v", pipeline1.Name, st),
			Message: PipelineJobMessage{
				AttributesJson: string(attrs_json),
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, pj.ID)

		assert.Equal(t, Waiting, pj.Status)
		assert.Equal(t, 0, len(dummyPublisher.Invocations))
	}

	// Publish Job Message soon when the pipeline is Opened
	for _, st := range []Status{Opened} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &PipelineJob{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-publishing-%v", pipeline1.Name, st),
			Message: PipelineJobMessage{
				AttributesJson: string(attrs_json),
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, pj.ID)

		assert.Equal(t, Published, pj.Status)
		assert.Equal(t, 1, len(dummyPublisher.Invocations))
		dummyPublisher.Invocations = []*PublishInvocation{}
	}

	// Raise error when create PipelineJob
	for _, st := range []Status{Broken, Closing, Closing_error, Closed} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &PipelineJob{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-waiting-%v", pipeline1.Name, st),
			Message: PipelineJobMessage{
				AttributesJson: string(attrs_json),
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.Error(t, err)
		assert.Empty(t, pj.ID)
		assert.Equal(t, 0, len(dummyPublisher.Invocations))
	}
}
