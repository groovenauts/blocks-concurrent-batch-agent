package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gae_support"
	"models"
	"test_utils"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

const (
	test_proj1  = "proj-123"
	test_proj2  = "proj-777"
	auth_header = "Authorization"
)

func TestActions(t *testing.T) {
	Setup(echo.New())

	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	h := &handler{}

	invalid_get_test := func(setup func(req *http.Request)) {
		req, err := inst.NewRequest(echo.GET, "/pipelines", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		setup(req)
		assert.NoError(t, err)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/pipelines")

		f := h.withAuth(h.index)
		if assert.NoError(t, f(c)) {
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

	req, err := inst.NewRequest(echo.GET, "/pipelines", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)
	auth, err := models.GlobalAuthAccessor.Create(ctx)
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
	req, err = inst.NewRequest(echo.POST, "/pipelines", strings.NewReader(json1))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/pipelines")

	f := h.withAuth(h.create)
	assert.NoError(t, f(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	s := rec.Body.String()

	pl := models.Pipeline{}
	if assert.NoError(t, json.Unmarshal([]byte(s), &pl)) {
		assert.Equal(t, test_proj1, pl.ProjectID)
		assert.NotNil(t, pl.ID)
	}

	// Test for show
	path := "/pipelines/" + pl.ID
	req, err = inst.NewRequest(echo.GET, path, nil)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = h.withPipeline(h.withAuth, h.show)
	if assert.NoError(t, f(c)) {
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
		models.Initialized,
		models.Broken,
		models.Building,
		models.Opened,
		models.Closing_error,
		models.Closed,
	} {
		expections = append(expections, expection{
			status: st,
			result: map[string][]string{
				"deploying": []string{},
				"closing":   []string{},
			},
		})
	}

	expections = append(expections, expection{
		status: models.Deploying,
		result: map[string][]string{
			"deploying": []string{pl.ID},
			"closing":   []string{},
		},
	})
	expections = append(expections, expection{
		status: models.Closing,
		result: map[string][]string{
			"deploying": []string{},
			"closing":   []string{pl.ID},
		},
	})

	for _, expection := range expections {
		// Test for refresh
		pl.Status = expection.status
		// https://github.com/golang/appengine/blob/master/aetest/instance.go#L32-L46
		ctx = appengine.NewContext(req)
		err = pl.Update(ctx)
		assert.NoError(t, err)

		f = gae_support.With(h.refresh)

		test_utils.RetryWith(10, func() func() {
			req, err = inst.NewRequest(echo.GET, "/pipelines/refresh", nil)
			assert.NoError(t, err)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(auth_header, token)

			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec)
			c.SetPath("/pipelines/refresh")

			err := f(c)
			if err != nil {
				return func() { assert.NoError(t, err) }
			}
			if http.StatusOK != rec.Code {
				return func() { assert.Equal(t, http.StatusOK, rec.Code) }
			}
			s := rec.Body.String()
			res := map[string][]string{}
			err = json.Unmarshal([]byte(s), &res)
			if err != nil {
				return func() { assert.NoError(t, err) }
			}
			if !assert.ObjectsAreEqual(expection.result, res) {
				return func() { assert.Equal(t, expection.result, res) }
			}
			return nil
		})
	}

	// Test for index
	req, err = inst.NewRequest(echo.GET, "/pipelines", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/pipelines")

	f = h.withAuth(h.index)
	if assert.NoError(t, f(c)) {
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

	// /pipelines/subscriptions
	req, err = inst.NewRequest(echo.GET, "/pipelines/subscriptions", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/pipelines/subscriptions")

	f = h.withAuth(h.subscriptions)
	if assert.NoError(t, f(c)) {
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
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = h.withPipeline(h.withAuth, h.destroy)
	if assert.NoError(t, f(c)) {
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
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = h.withPipeline(h.withAuth, h.destroy)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pl2 := models.Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.ProjectID)
		}
	}

	// 2nd Test for show
	path = "/pipelines/" + pl.ID
	req, err = inst.NewRequest(echo.GET, path, nil)
	req.Header.Set(auth_header, token)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = h.withPipeline(h.withAuth, h.show)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
