package pipeline

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
)

const (
	test_proj1  = "proj-123"
	test_proj2  = "proj-777"
	auth_header = "Authorization"
)

func TestActions(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)
	defer inst.Close()

	e := echo.New()
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
	auth, err := CreateAuth(ctx)
	assert.NoError(t, err)
	token := "Bearer " + auth.Token

	// Test for create
	json1 := `{"project_id":"proj-123"}`
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

	pl := Pipeline{}
	if assert.NoError(t, json.Unmarshal([]byte(s), &pl)) {
		assert.Equal(t, test_proj1, pl.Props.ProjectID)
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

	f = h.withPipeline(h.show)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pl2 := Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.Props.ProjectID)
		}
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
		pls := []Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pls)) {
			assert.Equal(t, 1, len(pls))
		}
	}

	// Make pipeline deletable
	pl.Props.Status = closed
	// https://github.com/golang/appengine/blob/master/aetest/instance.go#L32-L46
	ctx = appengine.NewContext(req)
	err = pl.update(ctx)
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

	f = h.withPipeline(h.destroy)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pl2 := Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.Props.ProjectID)
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

	f = h.withPipeline(h.show)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
