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

func TestJobCRUD(t *testing.T) {
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

	kinds := []string{"Jobs", "Pipelines", "Organizations"}
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
			job := &Job{
				Pipeline:   pipeline,
				IdByClient: fmt.Sprintf("%v-job-%v", pipelineName, i),
				Status:     Published,
				Message: JobMessage{
					AttributeMap: map[string]string{
						"foo": fmt.Sprintf("%v", i),
					},
				},
			}
			err = job.Create(ctx)
			assert.NoError(t, err)
		}
	}

	jobs, err := GlobalJobAccessor.All(ctx)
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
	for _, st := range []Status{Uninitialized, Waiting, Reserved, Building, Deploying} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &Job{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-waiting-%v", pipeline1.Name, st),
			Message: JobMessage{
				AttributeMap: map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, pj.ID)

		assert.Equal(t, Ready, pj.Status)
		assert.Equal(t, 0, len(dummyPublisher.Invocations))

		saved, err := GlobalJobAccessor.Find(ctx, pj.ID)
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

		pj := &Job{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-publishing-%v", pipeline1.Name, st),
			Message: JobMessage{
				AttributeMap: map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}
		err := pj.CreateAndPublishIfPossible(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, pj.ID)
		assert.Equal(t, pj.ID, pj.Message.AttributeMap[JobIdKey])

		assert.Equal(t, Published, pj.Status)
		assert.Equal(t, 1, len(dummyPublisher.Invocations))
		dummyPublisher.Invocations = []*PublishInvocation{}

		saved, err := GlobalJobAccessor.Find(ctx, pj.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(saved.Message.AttributeEntries))
		entry0 := saved.Message.AttributeEntries[0]
		assert.Equal(t, "download_files", entry0.Name)
		assert.Equal(t, string(download_files_json), entry0.Value)
		entry1 := saved.Message.AttributeEntries[1]
		assert.Equal(t, JobIdKey, entry1.Name)
		assert.Equal(t, saved.ID, entry1.Value)
	}

	// Raise error when create Job
	for _, st := range []Status{Broken, Closing, ClosingError, Closed} {
		pipeline1.Status = st
		err = pipeline1.Update(ctx)
		assert.NoError(t, err)

		pj := &Job{
			Pipeline:   pipeline1,
			IdByClient: fmt.Sprintf("%s-job-waiting-%v", pipeline1.Name, st),
			Message: JobMessage{
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

func TestJobUpdateStatusIfGreaterThanBefore(t *testing.T) {
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

	org1 := &Organization{Name: "org1"}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	pipeline := &Pipeline{
		Organization: org1,
		Name:         "dummy-pipeline1",
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

	download_files := "gcs://bucket1/path/to/file1"
	download_files_json, err := json.Marshal(download_files)
	assert.NoError(t, err)

	pj := &Job{
		Pipeline:   pipeline,
		Status:     Ready,
		IdByClient: fmt.Sprintf("%s-job1", pipeline.Name),
		Message: JobMessage{
			AttributeMap: map[string]string{
				"download_files": string(download_files_json),
			},
		},
	}
	err = pj.Create(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, pj.ID)

	type Pattern struct {
		curSt      JobStatus
		completed  bool
		step       JobStep
		stepSt     JobStepStatus
		expectedSt JobStatus
	}

	patterns := []Pattern{}
	// Normal cases
	for _, st := range []JobStatus{Ready, Publishing, PublishError, Published, Executing} {
		patterns = append(patterns, []Pattern{
			{st, false, INITIALIZING, SUCCESS, Executing},
			{st, false, NACKSENDING, SUCCESS, Executing},
			{st, false, CANCELLING, SUCCESS, Failure},
			{st, true, ACKSENDING, SUCCESS, Success},
		}...)
	}
	patterns = append(patterns, []Pattern{
		{Failure, false, INITIALIZING, SUCCESS, Failure},
		{Failure, false, NACKSENDING, SUCCESS, Failure},
		{Failure, false, CANCELLING, SUCCESS, Failure},
		{Failure, true, ACKSENDING, SUCCESS, Success},

		{Success, false, INITIALIZING, SUCCESS, Success},
		{Success, false, NACKSENDING, SUCCESS, Success},
		{Success, false, CANCELLING, SUCCESS, Success},
		{Success, true, ACKSENDING, SUCCESS, Success},
	}...)

	// Abnormal cases
	for _, st := range []JobStatus{Ready, Publishing, PublishError, Published, Executing} {
		patterns = append(patterns, []Pattern{
			{st, false, INITIALIZING, FAILURE, Executing},
			{st, false, NACKSENDING, FAILURE, st},
			{st, false, CANCELLING, FAILURE, st},
			{st, false, ACKSENDING, FAILURE, st},
		}...)
	}
	for _, st := range []JobStatus{Failure, Success} {
		patterns = append(patterns, []Pattern{
			{st, false, INITIALIZING, FAILURE, st},
			{st, false, NACKSENDING, FAILURE, st},
			{st, false, CANCELLING, FAILURE, st},
			{st, false, ACKSENDING, FAILURE, st},
		}...)
	}

	for _, pat := range patterns {
		pj.Status = pat.curSt
		err := pj.Update(ctx)
		assert.NoError(t, err)
		err = pj.UpdateStatusIfGreaterThanBefore(ctx, pat.completed, pat.step, pat.stepSt)
		assert.NoError(t, err)
		if !assert.Equal(t, pat.expectedSt, pj.Status) {
			fmt.Printf("Expected was %v but is %v for [%v %v %v %v %v]\n", pat.expectedSt, pj.Status, pat.curSt, pat.completed, pat.step, pat.stepSt, pat.expectedSt)
		}
	}
}
