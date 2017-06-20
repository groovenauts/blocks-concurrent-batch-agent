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

	req, err := inst.NewRequest(echo.GET, "/admin/auths", nil)
	assert.NoError(t, err)
	ctx := appengine.NewContext(req)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/admin/auths")

	test_utils.ClearDatastore(t, ctx, "Auths")
	aetest.Login(user, req)
	log.Debugf(ctx, "user: %v\n", user)

	f := withFlash(h.index)
	err = f(c)
	if err != nil {
		log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
	}
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	req, err = inst.NewRequest(echo.POST, "/admin/auths", strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/admin/auths")

	test_utils.ExpectChange(t, ctx, "Auths", 1, func() {
		f = withFlash(h.create)
		err = f(c)
		if err != nil {
			log.Errorf(ctx, "%v Error: %v\n", c.Path(), err)
		}
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	req, err = inst.NewRequest(echo.GET, "/admin/auths", nil)
	assert.NoError(t, err)
	ctx = appengine.NewContext(req)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/admin/auths")

	f = withFlash(h.index)
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

	// disable
	path := "/admin/auths/" + auth.ID + "/disable"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(auth.ID)

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

	// destroy
	path = "/admin/auths/" + auth.ID + "/delete"
	req, err = inst.NewRequest(echo.POST, path, strings.NewReader(""))
	assert.NoError(t, err)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath(path)
	c.SetParamNames("id")
	c.SetParamValues(auth.ID)

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
