package pipeline

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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

func ExpectToHaveProps(t *testing.T, plp *PipelineProps) {
	if plp.ProjectID != proj {
		t.Fatalf("ProjectId is expected %v but it was %v", proj, plp.ProjectID)
	}
}

func retryWith(max int, impl func() (func()) )  {
	for i := 0; i < max + 1; i++ {
		f := impl()
		if f == nil { return }
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
	_, err = CreatePipeline(ctx, &PipelineProps{})
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
		"Command",
	}
	for _, field := range fields {
		err := detectErrorFor(errors, field)
		assert.NotNil(t, err)
		assert.Equal(t, "required", err.ActualTag())
	}

	// CreatePipeline valid
	plp := PipelineProps{
		Name:          "pipeline01",
		ProjectID:     proj,
		Zone:          "us-central1-f",
		SourceImage:   "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
		MachineType:   "f1-micro",
		TargetSize:    2,
		ContainerSize: 2,
		ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		Command:       "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
	}
	pl, err := CreatePipeline(ctx, &plp)
	assert.NoError(t, err)
	log.Debugf(ctx, "pl %v\n", pl)
	key, err := datastore.DecodeKey(pl.ID)
	assert.NoError(t, err)

	pl2 := &PipelineProps{}
	err = datastore.Get(ctx, key, pl2)
	assert.NoError(t, err)
	ExpectToHaveProps(t, pl2)

	// FindPipeline
	pl3, err := FindPipeline(ctx, pl.ID)
	assert.NoError(t, err)
	ExpectToHaveProps(t, &pl3.Props)

	// Update status
	pl.Props.Status = building
	err = pl.update(ctx)
	assert.NoError(t, err)

	// GetAllPipeline
	pls, err := GetAllPipelines(ctx)
	assert.NoError(t, err)
	if len(pls) != 1 {
		t.Fatalf("len(pls) expects %v but was %v\n", 1, len(pls))
	}
	ExpectToHaveProps(t, &pls[0].Props)

	// Update status
	pl.Props.Status = opened
	err = pl.update(ctx)
	assert.NoError(t, err)

	// GetPipelineIDsByStatus
	statuses := []Status{
		initialized, broken, building, deploying,
		// opened,
		closing, closed,
	}
	for _, st := range statuses {
		retryWith(10, func() (func()) {
			keys, err := GetPipelineIDsByStatus(ctx, st)
			assert.NoError(t, err)
			if len(keys) == 0 {
				// OK
				return nil
			} else {
				// NG but retryWith calls this function only at the last time
				return func(){
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
		pl.Props.Status = st
		err = pl.destroy(ctx)
		if err == nil {
			t.Fatalf("Pipeline can't be destroyed with status %v\n", st)
		}
	}
	pl.Props.Status = closed
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
	assert.Equal(t, st, fmt.Sprintf(ft, closed))

	assert.Equal(t, "0", fmt.Sprintf(fv, initialized))
	assert.Equal(t, "1", fmt.Sprintf(fv, broken))
	assert.Equal(t, "2", fmt.Sprintf(fv, building))
	assert.Equal(t, "3", fmt.Sprintf(fv, deploying))
	assert.Equal(t, "4", fmt.Sprintf(fv, opened))
	assert.Equal(t, "5", fmt.Sprintf(fv, closing))
	assert.Equal(t, "6", fmt.Sprintf(fv, closed))
}
