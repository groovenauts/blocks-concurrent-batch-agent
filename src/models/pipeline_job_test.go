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
					AttributeMap: map[string]string{
						"foo": fmt.Sprintf("%v", i),
					},
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

	originalPublisher := GlobalPublisher
	dummyPublisher := &DummyPublisher{
		Invocations: []*PublishInvocation{},
	}
	GlobalPublisher = dummyPublisher

	defer func() {
		GlobalPublisher = originalPublisher
	}()

	// CreateAndPublishIfPossible
	download_files := "gcs://bucket1/path/to/file1"
	download_files_json, err := json.Marshal(download_files)
	assert.NoError(t, err)

	// Don't publish Job Message soon when the pipeline isn't Opened
	for _, st := range []Status{Uninitialized, Pending, Reserved, Building, Deploying} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &PipelineJob{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-waiting-%v", pipeline1.Name, st),
			Message: PipelineJobMessage{
				AttributeMap: map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, pj.ID)

		assert.Equal(t, Waiting, pj.Status)
		assert.Equal(t, 0, len(dummyPublisher.Invocations))

		saved, err := GlobalPipelineJobAccessor.Find(ctx, pj.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(saved.Message.AttributeEntries))
		entry1 := saved.Message.AttributeEntries[0]
		assert.Equal(t, "download_files", entry1.Name)
		assert.Equal(t, string(download_files_json), entry1.Value)
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
				AttributeMap: map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, pj.ID)
		assert.Equal(t, pj.ID, pj.Message.AttributeMap[PipelineJobIdKey])

		assert.Equal(t, Published, pj.Status)
		assert.Equal(t, 1, len(dummyPublisher.Invocations))
		dummyPublisher.Invocations = []*PublishInvocation{}

		saved, err := GlobalPipelineJobAccessor.Find(ctx, pj.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(saved.Message.AttributeEntries))
		entry0 := saved.Message.AttributeEntries[0]
		assert.Equal(t, "download_files", entry0.Name)
		assert.Equal(t, string(download_files_json), entry0.Value)
		entry1 := saved.Message.AttributeEntries[1]
		assert.Equal(t, PipelineJobIdKey, entry1.Name)
		assert.Equal(t, saved.ID, entry1.Value)
	}

	// Raise error when create PipelineJob
	for _, st := range []Status{Broken, Closing, ClosingError, Closed} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &PipelineJob{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-waiting-%v", pipeline1.Name, st),
			Message: PipelineJobMessage{
				AttributeMap: map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.Error(t, err)
		assert.Empty(t, pj.ID)
		assert.Equal(t, 0, len(dummyPublisher.Invocations))
	}
}
