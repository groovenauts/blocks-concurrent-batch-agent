package pipeline

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/go-playground/validator.v9"
)

func ClearDatastore(t *testing.T, ctx context.Context, kind string) {
	q := datastore.NewQuery(kind).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = datastore.DeleteMulti(ctx, keys); err != nil {
		t.Fatal(err)
	}
}

func ExpectChange(t *testing.T, ctx context.Context, kind string, diff int, f func()) {
	q0 := datastore.NewQuery(kind)
	c0, err := q0.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	f()
	q1 := datastore.NewQuery(kind)
	c1, err := q1.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if c1-c0 != diff {
		t.Fatalf("Expect diff is %v, but it changed from %v to %v in %v", diff, c0, c1, kind)
	}
}

const (
	proj = "test-project-x"
)

func ExpectToHaveProps(t *testing.T, pl *Pipeline) {
	if pl.ProjectID != proj {
		t.Fatalf("ProjectId is expected %v but it was %v", proj, pl.ProjectID)
	}
}

// retry for datastore's eventual consistency
func retryWith(max int, impl func() func()) {
	for i := 0; i < max+1; i++ {
		f := impl()
		if f == nil {
			return
		}
		if i == max {
			f()
		} else {
			// Exponential backoff
			d := time.Duration(math.Pow(2.0, float64(i)) * 5.0)
			time.Sleep(d * time.Millisecond)
		}
	}
}

func TestWatcherCalcDifferences(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.NoError(t, err)
	defer done()

	ClearDatastore(t, ctx, "Pipelines")

	detectErrorFor := func(errors validator.ValidationErrors, field string) validator.FieldError {
		for _, err := range errors {
			if err.StructField() == field {
				return err
			}
		}
		return nil
	}

	// CreatePipeline invalid
	err = CreatePipeline(ctx, &Pipeline{})
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

	// CreatePipeline valid
	pl := Pipeline{
		Name:      "pipeline01",
		ProjectID: proj,
		Zone:      "us-central1-f",
		BootDisk: PipelineVmDisk{
			SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		},
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
	}
	err = CreatePipeline(ctx, &pl)
	assert.NoError(t, err)
	log.Debugf(ctx, "pl %v\n", pl)
	key, err := datastore.DecodeKey(pl.ID)
	assert.NoError(t, err)

	pl2 := &Pipeline{}
	err = datastore.Get(ctx, key, pl2)
	assert.NoError(t, err)
	ExpectToHaveProps(t, pl2)

	// FindPipeline
	pl3, err := FindPipeline(ctx, pl.ID)
	assert.NoError(t, err)
	ExpectToHaveProps(t, pl3)

	// Update status
	pl.Status = building
	err = pl.update(ctx)
	assert.NoError(t, err)

	// GetAllPipeline
	pls, err := GetAllPipelines(ctx)
	assert.NoError(t, err)
	if len(pls) != 1 {
		t.Fatalf("len(pls) expects %v but was %v\n", 1, len(pls))
	}
	ExpectToHaveProps(t, pls[0])

	// Update status
	pl.Status = opened
	err = pl.update(ctx)
	assert.NoError(t, err)

	// GetPipelineIDsByStatus
	statuses := []Status{
		initialized, broken, building, deploying,
		// opened,
		closing, closed,
	}
	for _, st := range statuses {
		retryWith(10, func() func() {
			keys, err := GetPipelineIDsByStatus(ctx, st)
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

	keys, err := GetPipelineIDsByStatus(ctx, opened)
	assert.NoError(t, err)
	if len(keys) != 1 {
		t.Fatalf("len(keys) for opened expects %v but was %v\n", 1, len(keys))
	}
	if keys[0] != pl.ID {
		t.Fatalf("keys[0] expects %v but was %v\n", pl.ID, keys[0])
	}

	// destroy
	indestructible_statuses := []Status{
		initialized, broken, building, deploying, opened, closing,
		//closed,
	}
	for _, st := range indestructible_statuses {
		pl.Status = st
		err = pl.destroy(ctx)
		if err == nil {
			t.Fatalf("Pipeline can't be destroyed with status %v\n", st)
		}
	}
	pl.Status = closed
	err = pl.destroy(ctx)
	assert.NoError(t, err)
}

func TestStatusTypeAndValue(t *testing.T) {
	ft := "%T"
	fv := "%#v"
	st := "pipeline.Status"
	assert.Equal(t, st, fmt.Sprintf(ft, initialized))
	assert.Equal(t, st, fmt.Sprintf(ft, broken))
	assert.Equal(t, st, fmt.Sprintf(ft, building))
	assert.Equal(t, st, fmt.Sprintf(ft, deploying))
	assert.Equal(t, st, fmt.Sprintf(ft, opened))
	assert.Equal(t, st, fmt.Sprintf(ft, closing))
	assert.Equal(t, st, fmt.Sprintf(ft, closing_error))
	assert.Equal(t, st, fmt.Sprintf(ft, closed))

	assert.Equal(t, "0", fmt.Sprintf(fv, initialized))
	assert.Equal(t, "1", fmt.Sprintf(fv, broken))
	assert.Equal(t, "2", fmt.Sprintf(fv, building))
	assert.Equal(t, "3", fmt.Sprintf(fv, deploying))
	assert.Equal(t, "4", fmt.Sprintf(fv, opened))
	assert.Equal(t, "5", fmt.Sprintf(fv, closing))
	assert.Equal(t, "6", fmt.Sprintf(fv, closing_error))
	assert.Equal(t, "7", fmt.Sprintf(fv, closed))
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

	for st, name := range StatusStrings {
		pl := &Pipeline{
			Name:      "pipeline-" + name,
			ProjectID: proj,
			Zone:      "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:   "f1-micro",
			TargetSize:    2,
			ContainerSize: 2,
			ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
			Command:       "",
			Status:        st,
		}
		assert.NoError(t, CreatePipeline(ctx, pl))
		pipelines[st] = pl
	}

	res, err := GetActiveSubscriptions(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	subscription := res[0]
	assert.Equal(t, pipelines[opened].ID, subscription.PipelineID)
	assert.Equal(t, "pipeline-opened", subscription.Pipeline)
	assert.Equal(t, "projects/test-project-x/subscriptions/pipeline-opened-progress-subscription", subscription.Name)
}
