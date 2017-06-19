package admin

import (
	"html/template"

	"github.com/labstack/echo"
)

var e *echo.Echo

func Setup(echo *echo.Echo, dir string) *AuthHandler{
	e = echo

	auth := &AuthHandler{
		Views: &HandlerViews{
			templates: template.Must(template.ParseGlob(dir + "/auths/*.html")),
		},
	}

	g := e.Group("/admin/auths")
	g.GET("", withFlash(auth.index))
	g.POST("", withFlash(auth.create))
	g.POST("/:id/disable", auth.AuthHandler(auth.disable))
	g.POST("/:id/delete", auth.AuthHandler(auth.destroy))

	return auth
}
