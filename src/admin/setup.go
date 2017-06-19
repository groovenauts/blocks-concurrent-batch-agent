package admin

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

var e *echo.Echo

func Setup(echo *echo.Echo, dir string) {
	e = echo

	auth := &AuthHandler{}
	t := &Template{
		templates: template.Must(template.ParseGlob(dir + "/*.html")),
	}
	e.Renderer = t

	g := e.Group("/admin/auths")
	g.GET("", withFlash(auth.index))
	g.POST("", withFlash(auth.create))
	g.POST("/:id/disable", auth.AuthHandler(auth.disable))
	g.POST("/:id/delete", auth.AuthHandler(auth.destroy))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
