package pipeline

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	// "golang.org/x/net/context"
	// "google.golang.org/appengine"
	// "google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
	"google.golang.org/appengine/aetest"
	// "appengine_internal"
)

const (
	test_proj1 = "proj-123"
	test_proj2 = "proj-777"
)

func TestActions(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	FatalIfError(t, err)
	defer inst.Close()

	e := echo.New()
	h := &handler{}

	// Test for create
	json1 := `{"project_id":"proj-123"}`
	req, err := inst.NewRequest(echo.POST, "/pipelines", strings.NewReader(json1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	FatalIfError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/pipelines")

	f := withAEContext(h.create)
	assert.NoError(t, f(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	s := rec.Body.String()
	fmt.Printf("rec.Body: %v\n", s)

	pl := Pipeline{}
	if assert.NoError(t, json.Unmarshal([]byte(s), &pl)) {
		fmt.Printf("pl %v\n", pl)
		fmt.Printf("pl.props %v\n", pl.Props)
		assert.Equal(t, test_proj1, pl.Props.ProjectID)
		assert.NotNil(t, pl.ID)
	}

	// Test for show
	path := "/pipelines/" + pl.ID
	fmt.Printf("path to show: %v\n", path)
	req, err = inst.NewRequest(echo.GET, path, nil)
	FatalIfError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = withPipeline(h.show)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		fmt.Printf("rec.Body: %v\n", s)

		pl2 := Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.Props.ProjectID)
			fmt.Printf("pl: %v\n", pl2)
		}
	}

	// Test for index
	req, err = inst.NewRequest(echo.GET, "/pipelines", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	FatalIfError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/pipelines")

	f = withAEContext(h.index)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		fmt.Printf("rec.Body: %v\n", s)

		pls := []Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pls)) {
			assert.Equal(t, 1, len(pls))
			fmt.Printf("pls[0] %v\n", pls[0])
		}
	}

	// Test for close_task
	close_task_path := path + "/close_task"
	req, err = inst.NewRequest(echo.POST, close_task_path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	FatalIfError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(close_task_path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = pipelineTask("close")
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		fmt.Printf("rec.Body: %v\n", s)

		pl2 := Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.Props.ProjectID)
			assert.Equal(t, Status(closed), pl2.Props.Status)
		}
	}

	// Test for destroy
	req, err = inst.NewRequest(echo.DELETE, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	FatalIfError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = withPipeline(h.destroy)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		fmt.Printf("rec.Body: %v\n", s)

		pl2 := Pipeline{}
		if assert.NoError(t, json.Unmarshal([]byte(s), &pl2)) {
			assert.Equal(t, test_proj1, pl2.Props.ProjectID)
			fmt.Printf("pl: %v\n", pl2)
		}
	}

	// 2nd Test for show
	path = "/pipelines/" + pl.ID
	req, err = inst.NewRequest(echo.GET, path, nil)
	FatalIfError(t, err)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(pl.ID)

	f = withPipeline(h.show)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		s := rec.Body.String()
		fmt.Printf("rec.Body: %v\n", s)
	}
}
