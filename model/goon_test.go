package model

import (
	"testing"
	// "golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	// "google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/stretchr/testify/assert"
	// "google.golang.org/api/deploymentmanager/v2"
)

func TestGoonUsage(t *testing.T) {
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

	org := &Organization{Name: "org1"}
	orgKey, err := org.Create(ctx)
	assert.NoError(t, err)
	assert.Equal(t, orgKey.IntID(), org.ID)

	pbStore := &PipelineBaseStore{ParentKey: orgKey}
	pb := &PipelineBase{
		ParentKey: orgKey,
		Name:      "pipeline-base1",
		ProjectID: "dummy-proj-999",
		Zone:      "asia-northeast1-a",
		InstanceGroup: InstanceGroupBody{
			BootDisk: InstanceGroupVMDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
			},
			MachineType:           "f1-micro",
			InstanceSizeRequested: 1,
			StartupScript:         "",
		},
		Container: PipelineContainer{
			Name: "groovenauts/batch_type_iot_example:0.3.1",
			Size: 1,
		},
		HibernationDelay: 180,
		Status:           OpeningStarting,
	}

	pbKey, err := pbStore.Create(ctx, pb)
	assert.NoError(t, err)
	assert.NotNil(t, pbKey)

	jobStore := JobStore{ParentKey: pbKey}
	job := &Job{
		ParentKey:  pbKey,
		IDByClient: "testJob1",
		Status:     Inactive,
	}

	jobKey, err := jobStore.Create(ctx, job)
	assert.NoError(t, err)
	assert.NotNil(t, jobKey)

	log.Debugf(ctx, "orgKey: %v orgKey.IntID() => %v", orgKey, orgKey.IntID())
	log.Debugf(ctx, "pbKey: %v pbKey.IntID() => %v", pbKey, pbKey.IntID())
	log.Debugf(ctx, "jobKey: %v jobKey.IntID() => %v", jobKey, jobKey.IntID())

	pb2, err := pbStore.ByID(ctx, job.ParentKey.StringID())
	assert.NoError(t, err)
	assert.NotNil(t, pb2)
	assert.Equal(t, pb.Name, pb2.Name)
	assert.Equal(t, pb.ParentKey, pb2.ParentKey)
}
