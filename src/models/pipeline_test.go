package models

import (
	"fmt"
	"testing"
	"time"

	"test_utils"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
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

	// CreatePipeline valid
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
	err = pl.Create(ctx)
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
	assert.Equal(t, Initialized, pl.Status)

	pendingPl := &Pipeline{
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
	err = pendingPl.Create(ctx)
	assert.NoError(t, err)
	assert.Equal(t, Pending, pendingPl.Status)

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
		if i.Status == Initialized {
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
		Initialized, Broken, Building, Deploying,
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
		Initialized, Broken, Building, Deploying, Opened, Closing,
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
	assert.Equal(t, st, fmt.Sprintf(ft, Initialized))
	assert.Equal(t, st, fmt.Sprintf(ft, Broken))
	assert.Equal(t, st, fmt.Sprintf(ft, Pending))
	assert.Equal(t, st, fmt.Sprintf(ft, Building))
	assert.Equal(t, st, fmt.Sprintf(ft, Deploying))
	assert.Equal(t, st, fmt.Sprintf(ft, Opened))
	assert.Equal(t, st, fmt.Sprintf(ft, Closing))
	assert.Equal(t, st, fmt.Sprintf(ft, Closing_error))
	assert.Equal(t, st, fmt.Sprintf(ft, Closed))

	assert.Equal(t, "0", fmt.Sprintf(fv, Initialized))
	assert.Equal(t, "1", fmt.Sprintf(fv, Broken))
	assert.Equal(t, "2", fmt.Sprintf(fv, Pending))
	assert.Equal(t, "3", fmt.Sprintf(fv, Building))
	assert.Equal(t, "4", fmt.Sprintf(fv, Deploying))
	assert.Equal(t, "5", fmt.Sprintf(fv, Opened))
	assert.Equal(t, "6", fmt.Sprintf(fv, Closing))
	assert.Equal(t, "7", fmt.Sprintf(fv, Closing_error))
	assert.Equal(t, "8", fmt.Sprintf(fv, Closed))
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
		TokenAmount: 20,
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

	res, err := GlobalPipelineAccessor.GetActiveSubscriptions(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	subscription := res[0]
	assert.Equal(t, pipelines[Opened].ID, subscription.PipelineID)
	assert.Equal(t, "pipeline-opened", subscription.Pipeline)
	assert.Equal(t, "projects/test-project-x/subscriptions/pipeline-opened-progress-subscription", subscription.Name)
}

func TestGetPendingPipelines(t *testing.T) {
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
		theTime := now.Add( time.Duration(-1 * (6 - i) * 10) * time.Minute)
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
			CreatedAt: theTime,
			UpdatedAt: theTime,
		}
		assert.NoError(t, pl.Create(ctx))
		pipelines = append(pipelines, pl)
	}
	
	// TokenAmount: 10
	// pipeline-1 {TokenConsumption: 5} 50 min ago Initialized
	// pipeline-2 {TokenConsumption: 4} 40 min ago Initialized
	// pipeline-3 {TokenConsumption: 3} 30 min ago Pending
	// pipeline-4 {TokenConsumption: 2} 20 min ago Pending
	// pipeline-5 {TokenConsumption: 1} 10 min ago Pending
	assert.Equal(t, Initialized, pipelines[0].Status)
	assert.Equal(t, Initialized, pipelines[1].Status)
	assert.Equal(t, Pending, pipelines[2].Status)
	assert.Equal(t, Pending, pipelines[3].Status)
	assert.Equal(t, Pending, pipelines[4].Status)
	
	res, err := org1.PipelineAccessor().GetPendings(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(res))

	names := []string{}
	for _, pl := range res {
		names = append(names, pl.Name)
	}
	assert.Equal(t, []string{"pipeline-3", "pipeline-4", "pipeline-5"}, names)
}
