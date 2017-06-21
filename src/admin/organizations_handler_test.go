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

func TestOrganizationsHandler(t *testing.T) {
	handlers := Setup(echo.New(), "../../app/concurrent-batch-agent/admin/views")
	h, ok := handlers["orgs"].(*OrganizationsHandler)
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

	// GET /admin/orgs (no organizations)
	req, err := inst.NewRequest(echo.GET, "/admin/orgs", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/admin/orgs")

	test_utils.ClearDatastore(t, ctx, "Organizations") // Once in this function
	aetest.Login(user, req)
	log.Debugf(ctx, "user: %v\n", user)

	f := withFlash(h.Index)
	err = f(c)
	if err != nil {
		log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// GET /admin/orgs/new
	req, err = inst.NewRequest(echo.GET, "/admin/orgs/new", nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/admin/orgs/new")

	f = withFlash(h.New)
	err = f(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// POST /admin/orgs with INVALID data
	formData := `name=&memo=`
	req, err = inst.NewRequest(echo.POST, "/admin/orgs", strings.NewReader(formData))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/admin/orgs")

	test_utils.ExpectChange(t, ctx, "Organizations", 0, func() {
		f = withFlash(h.Create)
		err = f(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// POST /admin/orgs with VALID data
	formData = "name=ORG1&memo=First Organization"
	req, err = inst.NewRequest(echo.POST, "/admin/orgs", strings.NewReader(formData))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/admin/orgs")

	test_utils.ExpectChange(t, ctx, "Organizations", 1, func() {
		f = withFlash(h.Create)
		err = f(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
	})

	// GET /admin/orgs with ORG1
	req, err = inst.NewRequest(echo.GET, "/admin/orgs", nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/admin/orgs")

	f = withFlash(h.Index)
	err = f(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	orgs, err := models.GlobalOrganizationAccessor.All(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(orgs))
	org1 := orgs[0]
	assert.NotEmpty(t, org1.ID)
	assert.Equal(t, "ORG1", org1.Name)
	assert.Equal(t, "First Organization", org1.Memo)

	// GET /admin/orgs/:id
	path := "/admin/orgs/" + org1.ID
	req, err = inst.NewRequest(echo.GET, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(org1.ID)

	f = h.Identified(h.Show)
	err = f(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// GET /admin/orgs/:id/edit
	path = "/admin/orgs/" + org1.ID + "/edit"
	req, err = inst.NewRequest(echo.GET, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(org1.ID)

	f = h.Identified(h.Edit)
	err = f(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// POST /admin/orgs/:id/update with INVALID data
	formData = "name=&memo=First Organization"
	path = "/admin/orgs/" + org1.ID + "/update"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(formData))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(org1.ID)

	test_utils.ExpectChange(t, ctx, "Organizations", 0, func() {
		f = h.Identified(h.Update)
		err = f(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	org1Reloaded, err := models.GlobalOrganizationAccessor.Find(ctx, org1.ID)
	assert.Equal(t, "ORG1", org1Reloaded.Name)
	assert.Equal(t, "First Organization", org1Reloaded.Memo)

	// POST /admin/orgs/:id/update with VALID data
	formData = "name=org1&memo=First Organization"
	path = "/admin/orgs/" + org1.ID + "/update"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(formData))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(org1.ID)

	test_utils.ExpectChange(t, ctx, "Organizations", 0, func() {
		f = h.Identified(h.Update)
		err = f(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
	})

	org1Updated, err := models.GlobalOrganizationAccessor.Find(ctx, org1.ID)
	assert.Equal(t, "org1", org1Updated.Name)
	assert.Equal(t, "First Organization", org1Updated.Memo)

	// POST /admin/orgs/:id/delete
	path = "/admin/orgs/" + org1.ID + "/delete"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(org1.ID)

	test_utils.ExpectChange(t, ctx, "Organizations", -1, func() {
		f = h.Identified(h.Destroy)
		err = f(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
	})
}
