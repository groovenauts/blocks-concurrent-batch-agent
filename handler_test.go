package pipeline

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/aetest"
)

const (
	test_proj1 = "proj-123"
	test_proj2 = "proj-777"
)

func TestActions(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	assert.NoError(t, err)
	defer inst.Close()

	e := echo.New()
	h := &handler{}

	// Test for create
	json1 := `{"project_id":"proj-123"}`
	req, err := inst.NewRequest(echo.POST, "/pipelines", strings.NewReader(json1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/pipelines")

	f := withAEContext(h.create)
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
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = withPipeline(h.show)
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
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/pipelines")

	f = withAEContext(h.index)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pls := []Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pls)) {
			assert.Equal(t, 1, len(pls))
		}
	}

	// Test for close_task
	close_task_path := path + "/close_task"
	req, err = inst.NewRequest(echo.POST, close_task_path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(close_task_path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = pipelineTask("close")
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		pl2 := Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.Props.ProjectID)
			assert.Equal(t, Status(closed), pl2.Props.Status)
		}
	}

	// Test for destroy
	req, err = inst.NewRequest(echo.DELETE, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = withPipeline(h.destroy)
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
	assert.NoError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = withPipeline(h.show)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}
