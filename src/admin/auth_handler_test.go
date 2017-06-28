package admin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"models"
	"test_utils"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func TestAdminHandler(t *testing.T) {
	handlers := Setup(echo.New(), "../../app/concurrent-batch-agent/admin/views")
	h, ok := handlers["auths"].(*AuthHandler)
	assert.True(t, ok)

	os.Setenv("BATCH_AGENT_HOSTNAME", "test.local")

	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	user := &user.User{
		Email:      "test@example.com",
		AuthDomain: "example.com",
		Admin:      true,
		ID:         "1",
	}

	req, err := inst.NewRequest(echo.GET, "/admin/orgs", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)

	test_utils.ClearDatastore(t, ctx, "Organizations")
	org := &models.Organization{Name: "ORG1"}
	org.Create(ctx)

	// GET /admin/orgs/:org_id/auths
	path := "/admin/orgs/" + org.ID + "/auths"
	req, err = inst.NewRequest(echo.GET, path, nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id")
	c.SetParamValues(org.ID)

	test_utils.ClearDatastore(t, ctx, "Auths")
	aetest.Login(user, req)
	log.Debugf(ctx, "user: %v\n", user)

	f := h.WithOrg(h.index)
	err = f(c)
	if err != nil {
		log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// POST /admin/orgs/:org_id/auths
	path = "/admin/orgs/" + org.ID + "/auths"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id")
	c.SetParamValues(org.ID)

	test_utils.ExpectChange(t, ctx, "Auths", 1, func() {
		f = h.WithOrg(h.create)
		err = f(c)
		if err != nil {
			log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
		}
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// GET /admin/orgs/:org_id/auths
	path = "/admin/orgs/" + org.ID + "/auths"
	req, err = inst.NewRequest(echo.GET, path, nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id")
	c.SetParamValues(org.ID)

	f = h.WithOrg(h.index)
	err = f(c)
	if err != nil {
		log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	auths, err := models.GlobalAuthAccessor.GetAll(ctx)
	assert.NoError(t, err)
	auth := auths[0]
	assert.NotEmpty(t, auth.ID)

	// POST /admin/orgs/:org_id/auths/:id/disable
	path = "/admin/orgs/" + org.ID + "/auths/" + auth.ID + "/disable"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id", "id")
	c.SetParamValues(org.ID, auth.ID)

	f = h.Identified(h.disable)
	err = f(c)
	if err != nil {
		log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)

	log.Debugf(ctx, "auth: %q %v\n", auth.ID, auth)
	updated, err := models.GlobalAuthAccessor.Find(ctx, auth.ID)
	assert.NoError(t, err)
	assert.True(t, updated.Disabled)

	// POST /admin/orgs/:org_id/auths/:id/delete
	path = "/admin/orgs/" + org.ID + "/auths/" + auth.ID + "/delete"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("org_id", "id")
	c.SetParamValues(org.ID, auth.ID)

	test_utils.ExpectChange(t, ctx, "Auths", -1, func() {
		f = h.Identified(h.destroy)
		err = f(c)
		if err != nil {
			log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
		}
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
	})

}
