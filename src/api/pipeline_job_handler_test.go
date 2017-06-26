package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"models"
	"test_utils"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func TestPipelineJobHandlerActions(t *testing.T) {
	handlers := SetupRoutes(echo.New())

	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	h, ok := handlers["pipeline_jobs"].(*PipelineJobHandler)
	assert.True(t, ok)
	actions := h.buildActions()

	req, err := inst.NewRequest(echo.GET, "/", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)

	kinds := []string{"PipelineJobs", "Pipelines", "Organizations"}
	for _, k := range kinds {
		test_utils.ClearDatastore(t, ctx, k)
	}

	org1 := &models.Organization{Name: "org1"}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	pipelines := map[string]*models.Pipeline{}
	pipelineNames := []string{"pipeline1", "pipeline2"}
	for _, pipelineName := range pipelineNames {
		pipeline:= &models.Pipeline{
			Organization: org1,
			Name: pipelineName,
			ProjectID: "dummy-proj-111",
			Zone: "asia-northeast1-a",
			BootDisk: models.PipelineVmDisk{
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
			job := &models.PipelineJob{
				Pipeline: pipeline,
				IdByClient: fmt.Sprintf("%v-job-%v", pipelineName, i),
				Status: models.Published,
				Message: models.PipelineJobMessage{
					AttributesJson: fmt.Sprintf(`{"foo":%v}`, i),
				},
			}
			err = job.Create(ctx)
			assert.NoError(t, err)
		}
	}

	pl1 := pipelines["pipeline1"]

	// Not authenticated
	invalid_get_test := func(setup func(req *http.Request)) {
		req, err := inst.NewRequest(echo.GET, "/pipelines/" + pl1.ID + "/jobs", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		assert.NoError(t, err)
		setup(req)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/pipelines/" + pl1.ID + "/jobs")
		c.SetParamNames("pipeline_id")
		c.SetParamValues(pl1.ID)

		if assert.NoError(t, actions["index"](c)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		}
	}

	invalid_get_test(func(req *http.Request) {
		// do nothing
	})
	auth_headers := []string{
		"",
		"Bearer ",
		"Bearer invalid-token",
		"Bearer invalid-token:123456789",
	}
	for _, v := range auth_headers {
		invalid_get_test(func(req *http.Request) {
			req.Header.Set(auth_header, v)
		})
	}

	// Authenticated
	req, err = inst.NewRequest(echo.GET, "/pipelines/" + pl1.ID + "/jobs", nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)

	auth := &models.Auth{Organization: org1}
	err = auth.Create(ctx)
	assert.NoError(t, err)
	token := "Bearer " + auth.Token

	// Test for create
	json1 := `{` +
		`"id_by_client":"` + pl1.Name + `-job-new1"` +
		`,"message":{` +
		`"attributes_jbon":"{\"download_files\":\"gcs://bucket1/path/to/file1\"}"` +
		`}` +
		`}`
	req, err = inst.NewRequest(echo.POST, "/pipelines/"+pl1.ID+"/jobs", strings.NewReader(json1))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/pipelines/" + pl1.ID + "/jobs")
	c.SetParamNames("pipeline_id")
	c.SetParamValues(pl1.ID)

	assert.NoError(t, actions["create"](c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	s := rec.Body.String()

	pj := models.PipelineJob{}
	pj.Pipeline = pl1
	if assert.NoError(t, json.Unmarshal([]byte(s), &pj)) {
		assert.NotNil(t, pj.ID)
	}

	// Test for show
	path := "/jobs/" + pj.ID
	req, err = inst.NewRequest(echo.GET, path, nil)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pj.ID)

	if assert.NoError(t, actions["show"](c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pj2 := models.Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pj2)) {
			assert.Equal(t, pj.ID, pj2.ID)
		}
	}

}
