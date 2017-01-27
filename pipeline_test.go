package pipeline

import (
	"fmt"
	"testing"

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
	pls, err := GetAllPipeline(ctx)
	assert.NoError(t, err)
	if len(pls) != 1 {
		t.Fatalf("len(pls) expects %v but was %v\n", 1, len(pls))
	}
	ExpectToHaveProps(t, &pls[0].Props)

	// Update status
	pl.Props.Status = opened
	err = pl.update(ctx)
	assert.NoError(t, err)

	// GetAllActivePipelineIDs
	keys, err := GetAllActivePipelineIDs(ctx)
	assert.NoError(t, err)
	if len(keys) != 1 {
		t.Fatalf("len(keys) expects %v but was %v\n", 1, len(keys))
	}
	if keys[0] != pl.ID {
		t.Fatalf("keys[0] expects %v but was %v\n", pl.ID, keys[0])
	}

	// destroy
	statuses := []Status{
		initialized, broken, building, opened, closing,
		//closed,
		resizing, updating, recreating,
	}
	for _, st := range statuses {
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
	assert.Equal(t, st, fmt.Sprintf(ft, opened))
	assert.Equal(t, st, fmt.Sprintf(ft, resizing))
	assert.Equal(t, st, fmt.Sprintf(ft, updating))
	assert.Equal(t, st, fmt.Sprintf(ft, recreating))
	assert.Equal(t, st, fmt.Sprintf(ft, closing))
	assert.Equal(t, st, fmt.Sprintf(ft, closed))

	assert.Equal(t, "0", fmt.Sprintf(fv, initialized))
	assert.Equal(t, "1", fmt.Sprintf(fv, broken))
	assert.Equal(t, "2", fmt.Sprintf(fv, building))
	assert.Equal(t, "3", fmt.Sprintf(fv, opened))
	assert.Equal(t, "4", fmt.Sprintf(fv, resizing))
	assert.Equal(t, "5", fmt.Sprintf(fv, updating))
	assert.Equal(t, "6", fmt.Sprintf(fv, recreating))
	assert.Equal(t, "7", fmt.Sprintf(fv, closing))
	assert.Equal(t, "8", fmt.Sprintf(fv, closed))
}
