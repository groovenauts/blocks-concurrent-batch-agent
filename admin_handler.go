package pipeline

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var	store = sessions.NewCookieStore([]byte("something-very-secret"))

type adminHandler struct{
}

func init() {
	h := &adminHandler{
	}

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

func (h *adminHandler) withFlash(impl func(c echo.Context, flash *Flash) error) func(c echo.Context) error {
	return withAEContext(func(c echo.Context) error {
		session, err := store.Get(c.Request(), "admin-session")
		if err != nil {
			ctx := c.Get("aecontext").(context.Context)
			log.Errorf(ctx, "Failed to get session: %v\n", err)
			return err
		}

		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			// HttpOnly: true,
		}
		session.Values["foo"] = "bar"

		ctx := c.Get("aecontext").(context.Context)
		log.Debugf(ctx, "Flash#set session: %v\n", session)

		f := &Flash{session: session}
		err = f.load(c)
		if err != nil {
			return err
		}
		err = impl(c, f)
		e2 := session.Save(c.Request(), c.Response().Writer)
		if e2 != nil {
			ctx := c.Get("aecontext").(context.Context)
			log.Errorf(ctx, "Failed to save data to session: %v\n", e2)
		}
		return err
	})
}

// GET http://localhost:8080/admin/auths

type IndexRes struct {
	Flash *Flash
	Auths []*Auth
}

func (h *adminHandler) index(c echo.Context, f *Flash) error {
	ctx := c.Get("aecontext").(context.Context)
	auths, err := GetAllAuth(ctx)
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	r := IndexRes{
		Auths: auths,
	}
	r.Flash = f
	return c.Render(http.StatusOK, "index", &r)
}

// POST http://localhost:8080/admin/auths

type CreateRes struct {
	Flash    *Flash
	Auth     *Auth
	Hostname string
}

func (h *adminHandler) create(c echo.Context, f *Flash) error {
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
	r.Flash = f
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

func (h *adminHandler) AuthHandler(f func(c echo.Context, flash *Flash, ctx context.Context, auth *Auth) error) func(c echo.Context) error {
	return h.withFlash(func(c echo.Context, flash *Flash) error {
		ctx := c.Get("aecontext").(context.Context)
		auth, err := FindAuth(ctx, c.Param("id"))
		if err == ErrNoSuchAuth {
			e2 := flash.set(c, "alert", fmt.Sprintf("Auth not found for id: %v", c.Param("id")))
			if e2 != nil {
				log.Errorf(ctx, "Error to set flash message: %q\n", e2)
				return e2
			}
			return c.Redirect(http.StatusFound, "/admin/auths")
		}
		if err != nil {
			e2 := flash.set(c, "alert", fmt.Sprintf("Failed to find Auth for id: %v error: %v", c.Param("id"), err))
			if e2 != nil {
				log.Errorf(ctx, "Error to set flash message: %q\n", e2)
				return e2
			}
			return c.Redirect(http.StatusFound, "/admin/auths")
		}
		return f(c, flash, ctx, auth)
	})
}

// PUT http://localhost:8080/admin/auths/:id
func (h *adminHandler) disable(c echo.Context, f *Flash, ctx context.Context, auth *Auth) error {
	auth.Disabled = true
	err := auth.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Auth: %v because of %v\n", auth, err)
		f.set(c, "alert", fmt.Sprintf("Failed to update Auth. id: %v error: %v", auth.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	f.set(c, "notice", fmt.Sprintf("Disabled the Auth successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusFound, "/admin/auths")
}

// DELETE http://localhost:8080/admin/auths/:id
func (h *adminHandler) destroy(c echo.Context, f *Flash, ctx context.Context, auth *Auth) error {
	err := auth.destroy(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to destroy Auth: %v because of %v\n", auth, err)
		f.set(c, "alert", fmt.Sprintf("Failed to destroy Auth. id: %v error: %v", auth.ID, err))
		return c.Redirect(http.StatusFound, "/admin/auths")
	}
	f.set(c, "notice", fmt.Sprintf("The Auth is deleted successfully. id: %v", auth.ID))
	return c.Redirect(http.StatusFound, "/admin/auths")
}
