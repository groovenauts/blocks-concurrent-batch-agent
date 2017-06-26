package models

import (
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
		pipeline:= &Pipeline{
			Organization: org1,
			Name: pipelineName,
			ProjectID: "dummy-proj-111",
			Zone: "asia-northeast1-a",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
			},
			MachineType: "f1-micro",
			TargetSize: 1,
			ContainerSize: 1,
			ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		}
		err = pipeline.Create(ctx)
		assert.NoError(t, err)
		pipelines[pipelineName] = pipeline

		for i := 1; i < 3; i++ {
			job := &PipelineJob{
				Pipeline: pipeline,
				IdByClient: fmt.Sprintf("%v-job-%v", pipelineName, i),
				Status: Published,
				Message: PipelineJobMessage{
					AttributesJson: fmt.Sprintf(`{"foo":%v}`, i),
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
}
