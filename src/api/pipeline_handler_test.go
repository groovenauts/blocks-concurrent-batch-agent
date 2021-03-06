package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/models"
	"github.com/groovenauts/blocks-concurrent-batch-server/src/test_utils"
)

const (
	test_proj1  = "proj-123"
	test_proj2  = "proj-777"
	auth_header = "Authorization"
)

func TestActions(t *testing.T) {
	handlers := SetupRoutes(echo.New())

	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	h, ok := handlers["pipelines"].(*PipelineHandler)
	assert.True(t, ok)

	req, err := inst.NewRequest(echo.GET, "/orgs", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)

	test_utils.ClearDatastore(t, ctx, "Organizations")
	org := &models.Organization{Name: "ORG1"}
	org.Create(ctx)

	invalid_get_test := func(setup func(req *http.Request)) {
		req, err := inst.NewRequest(echo.GET, "/orgs"+org.ID+"/pipelines", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		setup(req)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/orgs" + org.ID + "/pipelines")
		c.SetParamNames("org_id")
		c.SetParamValues(org.ID)

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

	req, err = inst.NewRequest(echo.GET, "/orgs"+org.ID+"/pipelines", nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)

	auth := &models.Auth{Organization: org}
	err = auth.Create(ctx)
	assert.NoError(t, err)
	token := "Bearer " + auth.Token

	// Test for create
	json1 := `{` +
		`"name":"pipeline01"` +
		`,"project_id":"proj-123"` +
		`,"zone":"us-central1-f"` +
		`,"boot_disk":{` +
		`"source_image":"https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0"` +
		`}` +
		`,"machine_type":"f1-micro"` +
		`,"preemptible":false` +
		`,"target_size":2` +
		`,"container_size":2` +
		`,"container_name":"groovenauts/batch_type_iot_example:0.3.1"` +
		`,"command":"bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}"` +
		`,"dryrun":true` +
		`}`
	req, err = inst.NewRequest(echo.POST, "/orgs"+org.ID+"/pipelines", strings.NewReader(json1))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/orgs" + org.ID + "/pipelines")
	c.SetParamNames("org_id")
	c.SetParamValues(org.ID)

	assert.NoError(t, h.collection(h.create)(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	s := rec.Body.String()

	pl := models.Pipeline{}
	pl.Organization = org
	if assert.NoError(t, json.Unmarshal([]byte(s), &pl)) {
		assert.Equal(t, test_proj1, pl.ProjectID)
		assert.NotNil(t, pl.ID)
	}

	// Test for show
	path := "/orgs" + org.ID + "/pipelines/" + pl.ID
	req, err = inst.NewRequest(echo.GET, path, nil)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id", "id")
	c.SetParamValues(org.ID, pl.ID)

	if assert.NoError(t, h.member(h.show)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pl2 := models.Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.ProjectID)
		}
	}

	type expection struct {
		status models.Status
		result map[string][]string
	}

	expections := []expection{}
	for _, st := range []models.Status{
		models.Uninitialized,
		models.Broken,
		models.Building,
		models.Opened,
		models.ClosingError,
		models.Closed,
	} {
		expections = append(expections, expection{
			status: st,
			result: map[string][]string{
				"ORG1-deploying": []string{},
				"ORG1-closing":   []string{},
			},
		})
	}

	expections = append(expections, expection{
		status: models.Deploying,
		result: map[string][]string{
			"ORG1-deploying": []string{pl.ID},
			"ORG1-closing":   []string{},
		},
	})
	expections = append(expections, expection{
		status: models.Closing,
		result: map[string][]string{
			"ORG1-deploying": []string{},
			"ORG1-closing":   []string{pl.ID},
		},
	})

	// Test for index
	req, err = inst.NewRequest(echo.GET, "/orgs"+org.ID+"/pipelines", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/orgs" + org.ID + "/pipelines")
	c.SetParamNames("org_id")
	c.SetParamValues(org.ID)

	if assert.NoError(t, h.collection(h.index)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pls := []models.Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pls)) {
			assert.Equal(t, 1, len(pls))
		}
	}

	// https://github.com/golang/appengine/blob/master/aetest/instance.go#L32-L46
	ctx = appengine.NewContext(req)

	// Make pipeline opened
	pl.Status = models.Opened
	err = pl.Update(ctx)
	assert.NoError(t, err)

	// /pipelines/orgs/:org_id/subscriptions
	req, err = inst.NewRequest(echo.GET, "/orgs"+org.ID+"/pipelines/subscriptions", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/orgs" + org.ID + "/pipelines/subscriptions")
	c.SetParamNames("org_id")
	c.SetParamValues(org.ID)

	if assert.NoError(t, h.collection(h.subscriptions)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		subscriptions := []models.Subscription{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &subscriptions)) {
			if assert.Equal(t, 1, len(subscriptions)) {
				sub := subscriptions[0]
				assert.Equal(t, "pipeline01", sub.Pipeline)
				assert.Equal(t, "projects/proj-123/subscriptions/pipeline01-progress-subscription", sub.Name)
			}
		}
	}

	// Test for destroy failure
	req, err = inst.NewRequest(echo.DELETE, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id", "id")
	c.SetParamValues(org.ID, pl.ID)

	if assert.NoError(t, h.member(h.destroy)(c)) {
		assert.Equal(t, http.StatusNotAcceptable, rec.Code) // http.StatusUnprocessableEntity
		s := rec.Body.String()
		assert.Regexp(t, "(?i)can't destroy", s)
		assert.Regexp(t, "(?i)opened", s)
		assert.Regexp(t, "(?i)close before delete", s)
	}

	// Make pipeline deletable
	pl.Status = models.Closed
	err = pl.Update(ctx)
	assert.NoError(t, err)

	// Test for destroy
	req, err = inst.NewRequest(echo.DELETE, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id", "id")
	c.SetParamValues(org.ID, pl.ID)

	if assert.NoError(t, h.member(h.destroy)(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pl2 := models.Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.ProjectID)
		}
	}

	// 2nd Test for show
	path = "/orgs" + org.ID + "/pipelines/" + pl.ID
	req, err = inst.NewRequest(echo.GET, path, nil)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id", "id")
	c.SetParamValues(org.ID, pl.ID)

	if assert.NoError(t, h.member(h.show)(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
