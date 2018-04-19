package models

import (
	"testing"

	"test_utils"

	"github.com/stretchr/testify/assert"
	// "google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	// "google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	// "gopkg.in/go-playground/validator.v9"
)

func TestPipelineInstanceSizeLogCreate(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.NoError(t, err)
	defer done()

	test_utils.ClearDatastore(t, ctx, "Pipelines")

	org1 := &Organization{
		Name:        "org01",
		TokenAmount: 10,
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	pl0 := &Pipeline{
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

	err = pl0.Create(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, pl0.ID)

	// Reload Pipeline without Organization
	pl1 := &Pipeline{ID: pl0.ID}
	err = pl1.Reload(ctx)
	assert.NoError(t, err)

	err = pl1.LogInstanceSizeWithError(ctx, "2018-04-19 19:56:46 +0900 JST", 1)
	assert.NoError(t, err)
}
