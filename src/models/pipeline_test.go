package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/test_utils"
)

const (
	proj = "test-project-x"
)

func ExpectToHaveProps(t *testing.T, pl *Pipeline) {
	if pl.ProjectID != proj {
		t.Fatalf("ProjectId is expected %v but it was %v", proj, pl.ProjectID)
	}
}

func TestWatcherCalcDifferences(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.NoError(t, err)
	defer done()

	test_utils.ClearDatastore(t, ctx, "Pipelines")

	detectErrorFor := func(errors validator.ValidationErrors, field string) validator.FieldError {
		for _, err := range errors {
			if err.StructField() == field {
				return err
			}
		}
		return nil
	}

	// CreatePipeline invalid
	empty_pl := &Pipeline{}
	err = empty_pl.Create(ctx)
	assert.Error(t, err)
	errors := err.(validator.ValidationErrors)

	fields := []string{
		"Name",
		"ProjectID",
		"Zone",
		"SourceImage",
		"MachineType",
		"TargetSize",
		"ContainerSize",
		"ContainerName",
	}
	for _, field := range fields {
		err := detectErrorFor(errors, field)
		if assert.NotNil(t, err) {
			assert.Equal(t, "required", err.ActualTag())
		}
	}

	org1 := &Organization{
		Name:        "org01",
		TokenAmount: 10,
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	// CreatePipeline (CreateWithReserveOrWait) valid
	pl := &Pipeline{
		Organization: org1,
		Name:         "pipeline01",
		ProjectID:    proj,
		Zone:         "us-central1-f",
		BootDisk: PipelineVmDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:      "f1-micro",
		TargetSize:       2,
		ContainerSize:    2,
		ContainerName:    "groovenauts/batch_type_iot_example:0.3.1",
		Command:          "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
		TokenConsumption: 2,
	}
	err = pl.CreateWithReserveOrWait(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, pl.ID)
	key, err := datastore.DecodeKey(pl.ID)
	assert.NoError(t, err)

	pl2 := &Pipeline{}
	err = datastore.Get(ctx, key, pl2)
	assert.NoError(t, err)
	ExpectToHaveProps(t, pl2)

	// FindPipeline
	pl3, err := GlobalPipelineAccessor.Find(ctx, pl.ID)
	assert.NoError(t, err)
	ExpectToHaveProps(t, pl3)

	// org1.TokenAmount got reduced
	orgReloaded, err := GlobalOrganizationAccessor.Find(ctx, org1.ID)
	assert.NoError(t, err)
	assert.Equal(t, org1.TokenAmount-pl.TokenConsumption, orgReloaded.TokenAmount)
	assert.Equal(t, Reserved, pl.Status)

	waitingPl := &Pipeline{
		Organization: org1,
		Name:         "pipeline01",
		ProjectID:    proj,
		Zone:         "us-central1-f",
		BootDisk: PipelineVmDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:      "f1-micro",
		TargetSize:       2,
		ContainerSize:    2,
		ContainerName:    "groovenauts/batch_type_iot_example:0.3.1",
		Command:          "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
		TokenConsumption: org1.TokenAmount - pl.TokenConsumption + 1,
	}
	err = waitingPl.CreateWithReserveOrWait(ctx)
	assert.NoError(t, err)
	assert.Equal(t, Waiting, waitingPl.Status)

	// Update status
	pl.Status = Building
	err = pl.Update(ctx)
	assert.NoError(t, err)

	// GetAllPipeline
	pls, err := GlobalPipelineAccessor.GetAll(ctx)
	assert.NoError(t, err)
	if len(pls) != 2 {
		t.Fatalf("len(pls) expects %v but was %v\n", 2, len(pls))
	}
	var pls0 *Pipeline
	for _, i := range pls {
		if i.Status == Reserved {
			pls0 = i
		}
	}
	assert.NotNil(t, pls0)
	ExpectToHaveProps(t, pls0)

	// Update status
	pl.Status = Opened
	err = pl.Update(ctx)
	assert.NoError(t, err)

	// GetPipelineIDsByStatus
	statuses := []Status{
		Uninitialized, Broken,
		// Waiting,
		Reserved, Building, Deploying,
		// Opened,
		Closing, Closed,
	}
	for _, st := range statuses {
		test_utils.RetryWith(10, func() func() {
			keys, err := GlobalPipelineAccessor.GetIDsByStatus(ctx, st)
			assert.NoError(t, err)
			if len(keys) == 0 {
				// OK
				return nil
			} else {
				// NG but retryWith calls this function only at the last time
				return func() {
					t.Fatalf("len(keys) of %v expects %v but was %v\n", st, 0, len(keys))
				}
			}
		})
	}

	keys, err := GlobalPipelineAccessor.GetIDsByStatus(ctx, Opened)
	assert.NoError(t, err)
	if len(keys) != 1 {
		t.Fatalf("len(keys) for opened expects %v but was %v\n", 1, len(keys))
	}
	if keys[0] != pl.ID {
		t.Fatalf("keys[0] expects %v but was %v\n", pl.ID, keys[0])
	}

	// destroy
	indestructible_statuses := []Status{
		Uninitialized, Broken, Waiting, Reserved, Building, Deploying, Opened, Closing,
		//Closed,
	}
	for _, st := range indestructible_statuses {
		pl.Status = st
		err = pl.Destroy(ctx)
		if err == nil {
			t.Fatalf("Pipeline can't be destroyed with status %v\n", st)
		}
	}
	pl.Status = Closed
	err = pl.Destroy(ctx)
	assert.NoError(t, err)
}

func TestStatusTypeAndValue(t *testing.T) {
	ft := "%T"
	fv := "%#v"
	st := "models.Status"
	assert.Equal(t, st, fmt.Sprintf(ft, Uninitialized))
	assert.Equal(t, st, fmt.Sprintf(ft, Broken))
	assert.Equal(t, st, fmt.Sprintf(ft, Pending))
	assert.Equal(t, st, fmt.Sprintf(ft, Waiting))
	assert.Equal(t, st, fmt.Sprintf(ft, Reserved))
	assert.Equal(t, st, fmt.Sprintf(ft, Building))
	assert.Equal(t, st, fmt.Sprintf(ft, Deploying))
	assert.Equal(t, st, fmt.Sprintf(ft, Opened))
	assert.Equal(t, st, fmt.Sprintf(ft, HibernationChecking))
	assert.Equal(t, st, fmt.Sprintf(ft, HibernationStarting))
	assert.Equal(t, st, fmt.Sprintf(ft, HibernationProcessing))
	assert.Equal(t, st, fmt.Sprintf(ft, HibernationError))
	assert.Equal(t, st, fmt.Sprintf(ft, Hibernating))
	assert.Equal(t, st, fmt.Sprintf(ft, Closing))
	assert.Equal(t, st, fmt.Sprintf(ft, ClosingError))
	assert.Equal(t, st, fmt.Sprintf(ft, Closed))

	assert.Equal(t, "0", fmt.Sprintf(fv, Uninitialized))
	assert.Equal(t, "1", fmt.Sprintf(fv, Broken))
	assert.Equal(t, "2", fmt.Sprintf(fv, Pending))
	assert.Equal(t, "3", fmt.Sprintf(fv, Waiting))
	assert.Equal(t, "4", fmt.Sprintf(fv, Reserved))
	assert.Equal(t, "5", fmt.Sprintf(fv, Building))
	assert.Equal(t, "6", fmt.Sprintf(fv, Deploying))
	assert.Equal(t, "7", fmt.Sprintf(fv, Opened))
	assert.Equal(t, "8", fmt.Sprintf(fv, HibernationChecking))
	assert.Equal(t, "9", fmt.Sprintf(fv, HibernationStarting))
	assert.Equal(t, "10", fmt.Sprintf(fv, HibernationProcessing))
	assert.Equal(t, "11", fmt.Sprintf(fv, HibernationError))
	assert.Equal(t, "12", fmt.Sprintf(fv, Hibernating))
	assert.Equal(t, "13", fmt.Sprintf(fv, Closing))
	assert.Equal(t, "14", fmt.Sprintf(fv, ClosingError))
	assert.Equal(t, "15", fmt.Sprintf(fv, Closed))
}

func TestPipelineStateTransition(t *testing.T) {
	statuses := []Status{
		Uninitialized,
		Broken,
		Waiting,
		// Reserved,
		// Building,
		Deploying,
		Opened,
		Closing,
		ClosingError,
		Closed,
	}
	for _, st := range statuses {
		pl := &Pipeline{
			Organization: nil,
			Name:         "pipeline01",
			ProjectID:    "dummy-porj-999",
			Zone:         "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:      "f1-micro",
			TargetSize:       1,
			ContainerSize:    1,
			ContainerName:    "groovenauts/batch_type_iot_example:0.3.1",
			Command:          "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
			TokenConsumption: 1,
			Status:           st,
		}
		err := pl.StartBuilding(nil)
		assert.Error(t, err)
		_, ok := err.(*InvalidStateTransition)
		assert.True(t, ok)
	}
}

func TestGetActiveSubscriptions(t *testing.T) {
	// See https://github.com/golang/appengine/blob/master/aetest/instance.go#L36-L50
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

	pipelines := map[Status]*Pipeline{}

	org1 := &Organization{
		Name:        "org01",
		TokenAmount: len(StatusStrings) * 2,
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	for st, name := range StatusStrings {
		pl := &Pipeline{
			Organization: org1,
			Name:         "pipeline-" + name,
			ProjectID:    proj,
			Zone:         "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:      "f1-micro",
			TargetSize:       2,
			ContainerSize:    2,
			ContainerName:    "groovenauts/batch_type_iot_example:0.3.1",
			Command:          "",
			Status:           st,
			TokenConsumption: 2,
		}
		assert.NoError(t, pl.Create(ctx))
		pipelines[st] = pl
	}

	var res []*Subscription
	test_utils.RetryWith(12, func() func() {
		res, err = GlobalPipelineAccessor.GetActiveSubscriptions(ctx)
		assert.NoError(t, err)
		if 1 == len(res) {
			return nil
		} else {
			return func() {
				assert.Equal(t, 1, len(res))
			}
		}
	})

	subscription := res[0]
	assert.Equal(t, pipelines[Opened].ID, subscription.PipelineID)
	assert.Equal(t, "pipeline-opened", subscription.Pipeline)
	assert.Equal(t, "projects/test-project-x/subscriptions/pipeline-opened-progress-subscription", subscription.Name)
}

func TestGetWaitingPipelines(t *testing.T) {
	// See https://github.com/golang/appengine/blob/master/aetest/instance.go#L36-L50
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

	pipelines := []*Pipeline{}

	org1 := &Organization{
		Name:        "org01",
		TokenAmount: 10,
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	now := time.Now()
	for i := 1; i < 6; i++ {
		theTime := now.Add(time.Duration(-1*(6-i)*10) * time.Minute)
		pl := &Pipeline{
			Organization: org1,
			Name:         fmt.Sprintf("pipeline-%v", i),
			ProjectID:    proj,
			Zone:         "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:      "f1-micro",
			TargetSize:       6 - i,
			ContainerSize:    1,
			ContainerName:    "groovenauts/batch_type_iot_example:0.3.1",
			Command:          "",
			TokenConsumption: 6 - i,
			CreatedAt:        theTime,
			UpdatedAt:        theTime,
		}
		assert.NoError(t, pl.CreateWithReserveOrWait(ctx))
		pipelines = append(pipelines, pl)
	}

	// TokenAmount: 10
	// pipeline-1 {TokenConsumption: 5} 50 min ago Reserved
	// pipeline-2 {TokenConsumption: 4} 40 min ago Reserved
	// pipeline-3 {TokenConsumption: 3} 30 min ago Waiting
	// pipeline-4 {TokenConsumption: 2} 20 min ago Waiting
	// pipeline-5 {TokenConsumption: 1} 10 min ago Waiting
	assert.Equal(t, Reserved, pipelines[0].Status)
	assert.Equal(t, Reserved, pipelines[1].Status)
	assert.Equal(t, Waiting, pipelines[2].Status)
	assert.Equal(t, Waiting, pipelines[3].Status)
	assert.Equal(t, Waiting, pipelines[4].Status)

	res, err := org1.PipelineAccessor().GetWaitings(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(res))

	names := []string{}
	for _, pl := range res {
		names = append(names, pl.Name)
	}
	assert.Equal(t, []string{"pipeline-3", "pipeline-4", "pipeline-5"}, names)
}

func TestPipelineHasNewTaskSince(t *testing.T) {
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

	loc, err := time.LoadLocation("Local")
	assert.NoError(t, err)

	times := []time.Time{
		time.Date(2017, 11, 14, 0, 0, 0, 0, loc),
		time.Date(2017, 11, 14, 6, 0, 0, 0, loc),
		time.Date(2017, 11, 14, 12, 0, 0, 0, loc),
	}

	for idx, tt := range times {
		job := &Job{
			Pipeline:   pipeline,
			CreatedAt:  tt,
			Status:     Ready,
			IdByClient: fmt.Sprintf("%s-job%d", pipeline.Name, idx),
			Message: JobMessage{
				AttributeMap: map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}
		err = job.Create(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, job.ID)
	}

	patterns := map[time.Time]bool{
		times[0]:                      true,
		times[1]:                      true,
		times[2]:                      false,
		times[2].Add(1 * time.Second): false,
	}

	for tt, expected := range patterns {
		actual, err := pipeline.HasNewTaskSince(ctx, tt)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}
