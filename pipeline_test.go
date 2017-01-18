package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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

	// CreatePipeline
	pl, err := CreatePipeline(ctx, &PipelineProps{ProjectID: proj})
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

	// GetAllPipeline
	pls, err := GetAllPipeline(ctx)
	assert.NoError(t, err)
	if len(pls) != 1 {
		t.Fatalf("len(pls) expects %v but was %v\n", 1, len(pls))
	}
	ExpectToHaveProps(t, &pls[0].Props)

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
