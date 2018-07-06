package model

import (
	"testing"
	// "golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
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
	assert.NoError(t, org.Create(ctx))
	orgKey, err := datastore.DecodeKey(org.ID)
	assert.NoError(t, err)

	pbStore := &PipelineBaseStore{ParentKey: orgKey}
	pb := &PipelineBase{
		Parent: orgKey,
		Name: "pipeline-base1",
		ProjectID: "dummy-proj-999",
		Zone: "asia-northeast1-a",
		InstanceGroup: InstanceGroupBody{
			BootDisk: InstanceGroupVMDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
			},
			MachineType: "f1-micro",
			InstanceSizeRequested: 1,
			InstanceSize: 1,
			StartupScript: "",
		},
		Container: PipelineContainer{
			Name: "groovenauts/batch_type_iot_example:0.3.1",
			Size: 1,
		},
		HibernationDelay: 180,
		Status: OpeningStarting,
	}

	pbKey, err := pbStore.Create(ctx, pb)
	assert.NoError(t, err)
	assert.NotNil(t, pbKey)

	jobStore := JobStore{ParentKey: pbKey}
	job := &Job{
		Parent: pbKey,
		IDByClient: "testJob1",
		Status: Inactive,
		PipelineBaseId: pb.Id,
	}

	jobKey, err := jobStore.Create(ctx, job)
	assert.NoError(t, err)
	assert.NotNil(t, jobKey)

	log.Debugf(ctx, "orgKey: %v orgKey.StringID() => %v", orgKey, orgKey.StringID())
	log.Debugf(ctx, "pbKey: %v pbKey.StringID() => %v", pbKey, pbKey.StringID())
	log.Debugf(ctx, "jobKey: %v jobKey.StringID() => %v", jobKey, jobKey.StringID())

	pb2, err := pbStore.Get(ctx, job.Parent.StringID())
	assert.NoError(t, err)
	assert.NotNil(t, pb2)
	assert.Equal(t, pb.Id, pb2.Id)
}
