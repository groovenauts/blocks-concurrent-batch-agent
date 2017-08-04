package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	// "google.golang.org/appengine/log"
)

func TestDependencySatisfied(t *testing.T) {
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

	jobs := []*Job{}
	for i := 1; i < 4; i++ {
		job := &Job{
			Pipeline:   pipeline,
			IdByClient: fmt.Sprintf("dummy-pipeline1-job-%v", i),
			Status:     Ready,
			Message: JobMessage{
				AttributeMap: map[string]string{
					"foo": fmt.Sprintf("%v", i),
				},
			},
		}
		err = job.Create(ctx)
		assert.NoError(t, err)
		jobs = append(jobs, job)
	}

	jobIDs := []string{}
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	deps := map[string]*Dependency{
		"failure": &Dependency{Condition: OnFailure, JobIDs: jobIDs},
		"success": &Dependency{Condition: OnSuccess, JobIDs: jobIDs},
		"finish": &Dependency{Condition: OnFinish, JobIDs: jobIDs},
	}

	type Pattern struct {
		Statuses []JobStatus
		Results map[string]bool
	}

	allFalse := map[string]bool{"failure": false, "success": false, "finish": false}
	finishOnly := map[string]bool{"failure": false, "success": false, "finish": true}
	failureAndFinish := map[string]bool{"failure": true, "success": false, "finish": true}
	successAndFinish := map[string]bool{"failure": false, "success": true, "finish": true}

	falsePatterns := []Pattern{
		Pattern{Statuses: []JobStatus{Ready, Ready, Ready}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Ready, Ready, Publishing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Publishing, Publishing, Publishing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Publishing, PublishError, Publishing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{PublishError, PublishError, PublishError}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Published, Published, Published}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Published, Published, Executing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Published, Executing, Success}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Published, Executing, Failure}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Executing, Executing, Executing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Executing, Success, Executing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Success, Executing, Success}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Failure, Executing, Executing}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Failure, Executing, Failure}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Failure, Executing, Success}, Results: allFalse},
		Pattern{Statuses: []JobStatus{Failure, Success, Success}, Results: finishOnly},
		Pattern{Statuses: []JobStatus{Failure, Failure, Success}, Results: finishOnly},
		Pattern{Statuses: []JobStatus{Success, Success, Success}, Results: successAndFinish},
		Pattern{Statuses: []JobStatus{Failure, Failure, Failure}, Results: failureAndFinish},
	}

	for _, pat := range falsePatterns {
		for i, st := range pat.Statuses {
			job := jobs[i]
			job.Status = st
			err := job.Update(ctx)
			assert.NoError(t, err)
		}
		for k, expected := range pat.Results {
			dep := deps[k]
			r, err := dep.Satisfied(ctx)
			assert.NoError(t, err)
			assert.Equal(t, expected, r)
		}
	}
}
