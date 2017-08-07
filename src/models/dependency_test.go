package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
)

func SetupDependencyTest(t *testing.T, f func(context.Context, *Organization, *Pipeline, Jobs)) {
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

	jobs := Jobs{}
	for i := 1; i < 4; i++ {
		job := &Job{
			Pipeline:   pipeline,
			IdByClient: fmt.Sprintf("job-%v", i),
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
	f(ctx, org1, pipeline, jobs)
}

func TestDependencySatisfied(t *testing.T) {
	SetupDependencyTest(t, func(ctx context.Context, _ *Organization, pipeline *Pipeline, jobs Jobs) {

		jobIDs := jobs.IDs()

		deps := map[string]*Dependency{
			"failure": &Dependency{Condition: OnFailure, JobIDs: jobIDs},
			"success": &Dependency{Condition: OnSuccess, JobIDs: jobIDs},
			"finish":  &Dependency{Condition: OnFinish, JobIDs: jobIDs},
		}

		type Pattern struct {
			Statuses []JobStatus
			Results  map[string]bool
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
	})
}

func TestPipelineAccessorPendingsFor(t *testing.T) {
	SetupDependencyTest(t, func(ctx context.Context, org *Organization, pipeline *Pipeline, jobs Jobs) {
		jobIDs := jobs.IDs()

		type Pattern struct {
			Cond    DependencyCondition
			JobNums []int
		}

		patterns := map[string]Pattern{
			"F0": Pattern{Cond: OnFailure, JobNums: []int{0}},
			"F1": Pattern{Cond: OnFailure, JobNums: []int{0, 1}},
			"F2": Pattern{Cond: OnFailure, JobNums: []int{0, 1, 2}},
			"S0": Pattern{Cond: OnSuccess, JobNums: []int{0}},
			"S1": Pattern{Cond: OnSuccess, JobNums: []int{0, 1}},
			"S2": Pattern{Cond: OnSuccess, JobNums: []int{0, 1, 2}},
			"B0": Pattern{Cond: OnFinish, JobNums: []int{0}},
			"B1": Pattern{Cond: OnFinish, JobNums: []int{0, 1}},
			"B2": Pattern{Cond: OnFinish, JobNums: []int{0, 1, 2}},
		}

		pipelines := []*Pipeline{}

		for key, pat := range patterns {
			patJobIDs := []string{}
			for _, num := range pat.JobNums {
				patJobIDs = append(patJobIDs, jobIDs[num])
			}
			pipeline := &Pipeline{
				Organization: org,
				Name:         key,
				Status:       Pending,
				ProjectID:    "dummy-proj-111",
				Zone:         "asia-northeast1-a",
				BootDisk: PipelineVmDisk{
					SourceImage: "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
				},
				MachineType:   "f1-micro",
				TargetSize:    1,
				ContainerSize: 1,
				ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
				Dependency: Dependency{
					Condition: pat.Cond,
					JobIDs:    patJobIDs,
				},
			}
			err := pipeline.Create(ctx)
			assert.NoError(t, err)
			pipelines = append(pipelines, pipeline)
		}

		q1 := datastore.NewQuery("Pipelines").Filter("Dependency.JobIDs =", jobIDs[0])
		cnt1, err := q1.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 9, cnt1)

		q2 := datastore.NewQuery("Pipelines").Filter("Dependency.JobIDs =", jobIDs[2]).Filter("Dependency.Condition = ", OnFinish)
		cnt2, err := q2.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, cnt2)

		for i, st := range []JobStatus{Success, Executing, Published} {
			jobs[i].Status = st
			jobs[i].Update(ctx)
		}

		finished := jobs.Finished()
		assert.Equal(t, 1, len(finished))
		assert.Equal(t, "job-1", finished[0].IdByClient)

		pendingsTo2, err := GlobalPipelineAccessor.PendingsFor(ctx, []string{jobIDs[2]})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(pendingsTo2))
		for _, name := range []string{"F2", "S2", "B2"} {
			matched := false
			for _, pl := range pendingsTo2 {
				if name == pl.Name {
					matched = true
				}
			}
			assert.True(t, matched)
		}
	})
}
