package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/models"
	"github.com/groovenauts/blocks-concurrent-batch-server/src/test_utils"
)

func TestJobHandlerActions(t *testing.T) {
	handlers := SetupRoutes(echo.New())

	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	h, ok := handlers["jobs"].(*JobHandler)
	assert.True(t, ok)

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

		if assert.NoError(t, h.collection(h.index)(c)) {
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

	for num, st := range []models.JobStatus{models.Preparing, models.Ready} {
		job_id_base := fmt.Sprintf("%s-job-new-%d-1", pl1.Name, num)
		obj1 := map[string]interface{}{
			"id_by_client": fmt.Sprintf("%s-1", job_id_base),
			"message": map[string]interface{}{
				"attributes": map[string]string{
					"download_files": string(download_files_json),
				},
			},
		}

		json1, err := json.Marshal(obj1)
		assert.NoError(t, err)

		url := "/pipelines/" + pl1.ID + "/jobs"
		if st == models.Ready {
			url = url + "?ready=true"
		}
		req, err = inst.NewRequest(echo.POST, url, strings.NewReader(string(json1)))
		assert.NoError(t, err)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(auth_header, token)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/pipelines/" + pl1.ID + "/jobs")
		c.SetParamNames("pipeline_id")
		c.SetParamValues(pl1.ID)

		assert.NoError(t, h.collection(h.create)(c))
		assert.Equal(t, http.StatusCreated, rec.Code)

		s := rec.Body.String()

		job := models.Job{}
		job.Pipeline = pl1
		if assert.NoError(t, json.Unmarshal([]byte(s), &job)) {
			assert.NotNil(t, job.ID)
		}

		var jobRes map[string]interface{}
		assert.NoError(t, json.Unmarshal([]byte(s), &jobRes))
		assert.Equal(t, map[string]interface{}{
			"id":           job.ID,
			"id_by_client": fmt.Sprintf("%s-1", job_id_base),
			"status":       float64(st),
			"zone":         "",
			"hostname":     "",
			"published_at": "0001-01-01T00:00:00Z",
			"start_time":   "",
			"finish_time":  "",
			"message": map[string]interface{}{
				"attributes": map[string]interface{}{
					"download_files": `["gcs://bucket1/path/to/file1"]`,
				},
				"data": "",
			},
			"message_id": "",
			"created_at": job.CreatedAt.Format(time.RFC3339Nano),
			"updated_at": job.UpdatedAt.Format(time.RFC3339Nano),
		}, jobRes)

		// Test for invalid POST
		invalidAttrsPatterns := []interface{}{
			"VALID JSON String",
			[]string{"VALID JSON String Array"},
			map[string]int{"VALID JSON String to Integer": 1000},
		}
		for _, ptn := range invalidAttrsPatterns {
			obj := map[string]interface{}{
				"id_by_client": fmt.Sprintf("%s-2", job_id_base),
				"message": map[string]interface{}{
					"attributes": ptn,
				},
			}

			json2, err := json.Marshal(obj)
			assert.NoError(t, err)

			url := "/pipelines/" + pl1.ID + "/jobs"
			if st == models.Ready {
				url = url + "?ready=true"
			}

			req, err = inst.NewRequest(echo.POST, url, strings.NewReader(string(json2)))
			assert.NoError(t, err)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(auth_header, token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/pipelines/" + pl1.ID + "/jobs")
			c.SetParamNames("pipeline_id")
			c.SetParamValues(pl1.ID)

			assert.NoError(t, h.collection(h.create)(c))
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var res map[string]interface{}
			s := rec.Body.String()
			assert.NoError(t, json.Unmarshal([]byte(s), &res))
			assert.NotEmpty(t, res["error"])
		}

		// Test for show
		path := "/jobs/" + job.ID
		req, err = inst.NewRequest(echo.GET, path, nil)
		req.Header.Set(auth_header, token)
		assert.NoError(t, err)

		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("id")
		c.SetParamValues(job.ID)

		if assert.NoError(t, h.member(h.show)(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			s := rec.Body.String()
			job2 := models.Pipeline{}
			if assert.NoError(t, json.Unmarshal([]byte(s), &job2)) {
				assert.Equal(t, job.ID, job2.ID)
			}
		}

		// Test for bulk_get_jobs
		type BulkGetJobsPattern struct {
			jobId     string
			assertion func(map[string]interface{})
		}
		patterns := []BulkGetJobsPattern{
			BulkGetJobsPattern{
				jobId: job.ID,
				assertion: func(res map[string]interface{}) {
					assert.Empty(t, res["errors"])
					assert.NotEmpty(t, res["jobs"])
				},
			},
			BulkGetJobsPattern{
				jobId: "invalid-job-id",
				assertion: func(res map[string]interface{}) {
					assert.NotEmpty(t, res["errors"])
					assert.Empty(t, res["jobs"])
				},
			},
		}
		for _, ptn := range patterns {
			bulkGetJobsPayload1, err := json.Marshal(map[string][]string{
				"job_ids": []string{ptn.jobId},
			})
			assert.NoError(t, err)

			url := "/pipelines/" + pl1.ID + "/bulk_get_jobs"
			req, err = inst.NewRequest(echo.POST, url, strings.NewReader(string(bulkGetJobsPayload1)))
			assert.NoError(t, err)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(auth_header, token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath(url)
			c.SetParamNames("pipeline_id")
			c.SetParamValues(pl1.ID)

			assert.NoError(t, h.collection(h.BulkGetJobs)(c))
			assert.Equal(t, http.StatusOK, rec.Code)

			var res map[string]interface{}
			s := rec.Body.String()
			assert.NoError(t, json.Unmarshal([]byte(s), &res))
			ptn.assertion(res)
		}

		// POST /jobs/:id/getready
		path = "/jobs/" + job.ID + "/getready"
		req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
		req.Header.Set(auth_header, token)
		assert.NoError(t, err)

		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetPath(path)
		c.SetParamNames("id")
		c.SetParamValues(job.ID)

		assert.NoError(t, h.member(h.getReady)(c))
		assert.Equal(t, http.StatusOK, rec.Code)

		loaded, err := models.GlobalJobAccessor.Find(ctx, job.ID)
		assert.NoError(t, err)
		assert.Equal(t, models.Ready, loaded.Status)

		if st == models.Ready {
			// POST /jobs/:id/publish_task
			path = "/jobs/" + job.ID + "/publish_task"
			req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
			assert.NoError(t, err)
			req.Header.Set(auth_header, token)

			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec)
			c.SetPath(path)
			c.SetParamNames("id")
			c.SetParamValues(job.ID)

			msgId := "dummy-msg-id"
			// GlobalPublisher
			backup := models.GlobalPublisher
			models.GlobalPublisher = &DummyPublisher{ResultMessageId: msgId}
			defer (func() { models.GlobalPublisher = backup })()

			assert.NoError(t, h.member(h.PublishTask)(c))
			assert.Equal(t, http.StatusOK, rec.Code)

			reloaded, err := models.GlobalJobAccessor.Find(ctx, job.ID)
			assert.NoError(t, err)

			assert.Equal(t, msgId, reloaded.MessageID)
		}
	}

}
