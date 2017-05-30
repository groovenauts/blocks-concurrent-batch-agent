package pipeline

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type adminHandler struct{}

func init() {
	h := &adminHandler{}

	t := &Template{
		templates: template.Must(template.ParseGlob("admin/*.html")),
	}
	e.Renderer = t

	g := e.Group("/admin/auths")
	g.GET("", h.withFlash(h.index))
	g.POST("", h.withFlash(h.create))
	g.POST("/:id/disable", h.AuthHandler(h.disable))
	g.POST("/:id/delete", h.AuthHandler(h.destroy))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Flash struct {
	Alert  string
	Notice string
}

func (h *adminHandler) setFlash(c echo.Context, name, value string) {
	h.setFlashWithExpire(c, name, value, time.Now().Add(10*time.Minute))
}

func (h *adminHandler) setFlashWithExpire(c echo.Context, name, value string, expire time.Time) {
	cookie := new(http.Cookie)
	cookie.Path = "/admin/"
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = expire
	c.SetCookie(cookie)
}

func (h *adminHandler) loadFlash(c echo.Context) *Flash {
	f := Flash{}
	cookie, err := c.Cookie("alert")
	if err == nil {
		f.Alert = cookie.Value
	}
	cookie, err = c.Cookie("notice")
	if err == nil {
		f.Notice = cookie.Value
	}
	return &f
}

func (h *adminHandler) clearFlash(c echo.Context) {
	_, err := c.Cookie("alert")
	if err == nil {
		h.setFlashWithExpire(c, "alert", "", time.Now().AddDate(0, 0, 1))
	}
	_, err = c.Cookie("notice")
	if err == nil {
		h.setFlashWithExpire(c, "notice", "", time.Now().AddDate(0, 0, 1))
	}
}

func (h *adminHandler) withFlash(impl func(c echo.Context) error) func(c echo.Context) error {
	return withAEContext(func(c echo.Context) error {
		f := h.loadFlash(c)
		c.Set("flash", f)
		h.clearFlash(c)
		return impl(c)
	})
}

// GET http://localhost:8080/admin/auths

type IndexRes struct {
	Flash *Flash
	Auths []*Auth
}

func (h *adminHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	auths, err := GetAllAuth(ctx)
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	r := IndexRes{
		Auths: auths,
	}
	r.Flash = c.Get("flash").(*Flash)
	return c.Render(http.StatusOK, "index", &r)
}

// POST http://localhost:8080/admin/auths

type CreateRes struct {
	Flash    *Flash
	Auth     *Auth
	Hostname string
}

func (h *adminHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	auth, err := CreateAuth(ctx)
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
	return c.Render(http.StatusOK, "create", &r)
}

func (h *adminHandler) getHostname(c echo.Context) (string, error) {
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

func (h *adminHandler) AuthHandler(f func(c echo.Context, ctx context.Context, auth *Auth) error) func(c echo.Context) error {
	return h.withFlash(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		auth, err := FindAuth(ctx, c.Param("id"))
		if err == ErrNoSuchAuth {
			h.setFlash(c, "alert", fmt.Sprintf("Auth not found for id: %v", c.Param("id")))
			return c.Redirect(http.StatusFound, "/admin/auths")
		}
		if err != nil {
			h.setFlash(c, "alert", fmt.Sprintf("Failed to find Auth for id: %v error: %v", c.Param("id"), err))
			return c.Redirect(http.StatusFound, "/admin/auths")
		}
		return f(c, ctx, auth)
	})
}

// PUT http://localhost:8080/admin/auths/:id
func (h *adminHandler) disable(c echo.Context, ctx context.Context, auth *Auth) error {
	auth.Disabled = true
	err := auth.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Auth: %v because of %v\n", auth, err)
		h.setFlash(c, "alert", fmt.Sprintf("Failed to update Auth. id: %v error: %v", auth.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	h.setFlash(c, "notice", fmt.Sprintf("Disabled the Auth successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusFound, "/admin/auths")
}

// DELETE http://localhost:8080/admin/auths/:id
func (h *adminHandler) destroy(c echo.Context, ctx context.Context, auth *Auth) error {
	err := auth.destroy(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to destroy Auth: %v because of %v\n", auth, err)
		h.setFlash(c, "alert", fmt.Sprintf("Failed to destroy Auth. id: %v error: %v", auth.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	h.setFlash(c, "notice", fmt.Sprintf("The Auth is deleted successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusFound, "/admin/auths")
}
