package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"models"
	"test_utils"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func TestJobHandlerActions(t *testing.T) {
	handlers := SetupRoutes(echo.New())

	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	h, ok := handlers["jobs"].(*JobHandler)
	assert.True(t, ok)
	actions := h.buildActions()

	req, err := inst.NewRequest(echo.GET, "/", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)

	kinds := []string{"Jobs", "Pipelines", "Organizations"}
	for _, k := range kinds {
		test_utils.ClearDatastore(t, ctx, k)
	}

	org1 := &models.Organization{Name: "org1"}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	pipelines := map[string]*models.Pipeline{}
	pipelineNames := []string{"pipeline1", "pipeline2"}
	for _, pipelineName := range pipelineNames {
		pipeline := &models.Pipeline{
			Organization: org1,
			Name:         pipelineName,
			ProjectID:    "dummy-proj-111",
			Zone:         "asia-northeast1-a",
			BootDisk: models.PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable",
			},
			MachineType:   "f1-micro",
			TargetSize:    1,
			ContainerSize: 1,
			ContainerName: "groovenauts/batch_type_iot_example:0.3.1",
		}
		err = pipeline.Create(ctx)
		assert.NoError(t, err)
		pipelines[pipelineName] = pipeline

		for i := 1; i < 3; i++ {
			job := &models.Job{
				Pipeline:   pipeline,
				IdByClient: fmt.Sprintf("%v-job-%v", pipelineName, i),
				Status:     models.Published,
				Message: models.JobMessage{
					AttributeMap: map[string]string{
						"foo": fmt.Sprintf("%v", i),
					},
				},
			}
			err = job.Create(ctx)
			assert.NoError(t, err)
		}
	}

	pl1 := pipelines["pipeline1"]

	// Not authenticated
	invalid_get_test := func(setup func(req *http.Request)) {
		req, err := inst.NewRequest(echo.GET, "/pipelines/"+pl1.ID+"/jobs", nil)
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
	req, err = inst.NewRequest(echo.GET, "/pipelines/"+pl1.ID+"/jobs", nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)

	auth := &models.Auth{Organization: org1}
	err = auth.Create(ctx)
	assert.NoError(t, err)
	token := "Bearer " + auth.Token

	// Test for create
	download_files := []string{"gcs://bucket1/path/to/file1"}
	download_files_json, err := json.Marshal(download_files)
	assert.NoError(t, err)

	obj1 := map[string]interface{}{
		"id_by_client": pl1.Name + `-job-new1`,
		"message": map[string]interface{}{
			"attributes": map[string]string{
				"download_files": string(download_files_json),
			},
		},
	}

	json1, err := json.Marshal(obj1)
	assert.NoError(t, err)

	req, err = inst.NewRequest(echo.POST, "/pipelines/"+pl1.ID+"/jobs", strings.NewReader(string(json1)))
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

	pj := models.Job{}
	pj.Pipeline = pl1
	if assert.NoError(t, json.Unmarshal([]byte(s), &pj)) {
		assert.NotNil(t, pj.ID)
	}

	var pjRes map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(s), &pjRes))
	assert.Equal(t, map[string]interface{}{
		"id":           pj.ID,
		"id_by_client": "pipeline1-job-new1",
		"status":       float64(0),
		"message": map[string]interface{}{
			"attributes": map[string]interface{}{
				"download_files": `["gcs://bucket1/path/to/file1"]`,
			},
			"data": "",
		},
		"message_id": "",
		"created_at": pj.CreatedAt.Format(time.RFC3339Nano),
		"updated_at": pj.UpdatedAt.Format(time.RFC3339Nano),
	}, pjRes)

	// Test for invalid POST
	invalidAttrsPatterns := []interface{}{
		"VALID JSON String",
		[]string{"VALID JSON String Array"},
		map[string]int{"VALID JSON String to Integer": 1000},
	}
	for _, ptn := range invalidAttrsPatterns {
		obj := map[string]interface{}{
			"id_by_client": pl1.Name + `-job-new1"`,
			"message": map[string]interface{}{
				"attributes": ptn,
			},
		}

		json2, err := json.Marshal(obj)
		assert.NoError(t, err)

		req, err = inst.NewRequest(echo.POST, "/pipelines/"+pl1.ID+"/jobs", strings.NewReader(string(json2)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(auth_header, token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/pipelines/" + pl1.ID + "/jobs")
		c.SetParamNames("pipeline_id")
		c.SetParamValues(pl1.ID)

		assert.NoError(t, actions["create"](c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var res map[string]interface{}
		s := rec.Body.String()
		assert.NoError(t, json.Unmarshal([]byte(s), &res))
		assert.NotEmpty(t, res["error"])
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
