package pipeline

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"encoding/json"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	// "golang.org/x/net/context"
	// "google.golang.org/appengine"
	// "google.golang.org/appengine/log"
	"google.golang.org/appengine/aetest"
	// "appengine_internal"
)

const (
	test_proj1  = "proj-123"
	test_proj2  = "proj-777"
)

func TestActions(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	FatalIfError(t, err)
	defer inst.Close()

	json1 := `{"project_id":"proj-123"}`
	req, err := inst.NewRequest(echo.POST, "/pipelines", strings.NewReader(json1))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	FatalIfError(t, err)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/pipelines")
	h := &handler{}

	f := withAEContext(h.create)
	if assert.NoError(t, f(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// fmt.Printf("rec.Body: %v\n", rec.Body.String())

		pl := Pipeline{}
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pl)) {
			fmt.Printf("pl %v\n", pl)
			fmt.Printf("pl.props %v\n", pl.props)
			assert.Equal(t, test_proj1, pl.props.ProjectID)
			assert.NotNil(t, pl.id)
		}
	}
}
