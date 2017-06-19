package admin

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

var e *echo.Echo

func Setup(echo *echo.Echo, dir string) {
	e = echo

	h := &AuthHandler{}
	t := &Template{
		templates: template.Must(template.ParseGlob(dir + "/*.html")),
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
