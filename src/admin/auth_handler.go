package admin

import (
	"fmt"
	"net/http"
	"os"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type AuthHandler struct{
	Views Views
}


// GET http://localhost:8080/admin/auths

type IndexRes struct {
	Flash *Flash
	Auths []*models.Auth
}

func (h *AuthHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	auths, err := models.GlobalAuthAccessor.GetAll(ctx)
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	r := IndexRes{
		Auths: auths,
	}
	r.Flash = c.Get("flash").(*Flash)
	return h.Views.Render(c, http.StatusOK, "index", &r)
}

// POST http://localhost:8080/admin/auths

type CreateRes struct {
	Flash    *Flash
	Auth     *models.Auth
	Hostname string
}

func (h *AuthHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	auth := &models.Auth{}
	err := auth.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Error on create auth: %v\n", err)
		return err
	}
	hostname, err := h.getHostname(c)
	if err != nil {
		return err
	}
	r := CreateRes{
		Auth:     auth,
		Hostname: hostname,
	}
	r.Flash = c.Get("flash").(*Flash)
	return h.Views.Render(c, http.StatusOK, "create", &r)
}

func (h *AuthHandler) getHostname(c echo.Context) (string, error) {
	r := os.ExpandEnv("BATCH_AGENT_HOSTNAME")
	if r != "" {
		return r, nil
	}
	ctx := c.Get("aecontext").(context.Context)
	hostname, err := appengine.ModuleHostname(ctx, "", "", "")
	if err != nil {
		log.Errorf(ctx, "Failed to get ModuleHostname: %v\n", err)
		return "", err
	}
	return hostname, err
}

func (h *AuthHandler) AuthHandler(f func(c echo.Context, ctx context.Context, auth *models.Auth) error) func(c echo.Context) error {
	return withFlash(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		auth, err := models.GlobalAuthAccessor.Find(ctx, c.Param("id"))
		if err == models.ErrNoSuchAuth {
			setFlash(c, "alert", fmt.Sprintf("Auth not found for id: %v", c.Param("id")))
			return c.Redirect(http.StatusFound, "/admin/auths")
		}
		if err != nil {
			setFlash(c, "alert", fmt.Sprintf("Failed to find Auth for id: %v error: %v", c.Param("id"), err))
			return c.Redirect(http.StatusFound, "/admin/auths")
		}
		return f(c, ctx, auth)
	})
}

// PUT http://localhost:8080/admin/auths/:id
func (h *AuthHandler) disable(c echo.Context, ctx context.Context, auth *models.Auth) error {
	auth.Disabled = true
	err := auth.Update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Auth: %v because of %v\n", auth, err)
		setFlash(c, "alert", fmt.Sprintf("Failed to update Auth. id: %v error: %v", auth.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	setFlash(c, "notice", fmt.Sprintf("Disabled the Auth successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusFound, "/admin/auths")
}

// DELETE http://localhost:8080/admin/auths/:id
func (h *AuthHandler) destroy(c echo.Context, ctx context.Context, auth *models.Auth) error {
	err := auth.Destroy(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to destroy Auth: %v because of %v\n", auth, err)
		setFlash(c, "alert", fmt.Sprintf("Failed to destroy Auth. id: %v error: %v", auth.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	setFlash(c, "notice", fmt.Sprintf("The Auth is deleted successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusFound, "/admin/auths")
}
