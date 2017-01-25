package pipeline

import (
	"fmt"
	"io"
	"net/http"
	"html/template"
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
	g.GET(".html", withAEContext(h.index))
	g.POST(".html", withAEContext(h.create))
	g.PUT("/:id.html", h.AuthHandler(h.disable))
	g.DELETE("/:id.html", h.AuthHandler(h.destroy))
}

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type Flash struct {
	Alert string
	Notice string
}

func (h *adminHandler) setFlash(c echo.Context, name, value string) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Expires = time.Now().Add(10 * time.Minute)
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

// GET http://localhost:8080/admin/auths.html

type IndexRes struct {
	Flash *Flash
	Auths []*Auth
}

func (h *adminHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	log.Debugf(ctx, "index\n")
	auths, err := GetAllAuth(ctx)
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	log.Debugf(ctx, "indexPage auths: %v\n", auths)
	r := IndexRes{
		Auths: auths,
	}
	r.Flash = h.loadFlash(c)
	return c.Render(http.StatusOK, "index", &r)
}

// POST http://localhost:8080/admin/auths.html

type CreateRes struct {
	Flash *Flash
	Auth *Auth
	Hostname string
}

func (h *adminHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	log.Debugf(ctx, "create\n")
	auth, err := CreateAuth(ctx)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "create auth: %v\n", auth)
	hostname, err := appengine.ModuleHostname(ctx, "", "", "")
	if err != nil {
		return err
	}
	r := CreateRes{
		Auth: auth,
		Hostname: hostname,
	}
	r.Flash = h.loadFlash(c)
	return c.Render(http.StatusOK, "create", &r)
}

func (h *adminHandler) AuthHandler(f func(c echo.Context, ctx context.Context, auth *Auth) error) (func(c echo.Context) error) {
	return withAEContext(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		auth, err := FindAuth(ctx, c.Param("id"))
		if err == ErrNoSuchAuth {
			h.setFlash(c, "alert", fmt.Sprintf("Auth not found for id: %v", c.Param("id")))
			return c.Redirect(http.StatusNotModified, "/admin/auths.html")
		}
		if err != nil {
			h.setFlash(c, "alert", fmt.Sprintf("Failed to find Auth for id: %v error: ", c.Param("id"), err))
			return c.Redirect(http.StatusNotModified, "/admin/auths.html")
		}
		return f(c, ctx, auth)
	})
}

// PUT http://localhost:8080/admin/auths/:id.html
func (h *adminHandler) disable(c echo.Context, ctx context.Context, auth *Auth) error {
	auth.Props.Disabled = true
	err := auth.update(ctx)
	if err != nil {
		h.setFlash(c, "alert", fmt.Sprintf("Failed to update Auth. id: %v error: ", auth.ID, err))
		return c.Redirect(http.StatusNotModified, "/admin/auths.html")
	}
	h.setFlash(c, "notice", fmt.Sprintf("Disabled the Auth successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusSeeOther, "/admin/auths.html")
}

// DELETE http://localhost:8080/admin/auths/:id.html
func (h *adminHandler) destroy(c echo.Context, ctx context.Context, auth *Auth) error {
	err := auth.destroy(ctx)
	if err != nil {
		h.setFlash(c, "alert", fmt.Sprintf("Failed to destroy Auth. id: %v error: ", auth.ID, err))
		return c.Redirect(http.StatusNotModified, "/admin/auths.html")
	}
	h.setFlash(c, "notice", fmt.Sprintf("The Auth is disabled. id: %v", auth.ID))
	return c.Redirect(http.StatusSeeOther, "/admin/auths.html")
}