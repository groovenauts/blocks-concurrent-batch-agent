package admin

import (
	"html/template"

	"github.com/labstack/echo"
)

var e *echo.Echo

func Setup(echo *echo.Echo, dir string) map[string]interface{} {
	e = echo

	orgs := &OrganizationsHandler{
		Views: &HandlerViews{
			templates: template.Must(template.ParseGlob(dir + "/organizations/*.html")),
		},
	}

	gorgs := e.Group("/admin/orgs")
	gorgs.GET("", withFlash(orgs.Index))
	gorgs.GET("/new", withFlash(orgs.New))
	gorgs.POST("", withFlash(orgs.Create))
	gorgs.GET("/:id", orgs.Identified(orgs.Show))
	gorgs.GET("/:id/edit", orgs.Identified(orgs.Edit))
	gorgs.POST("/:id/update", orgs.Identified(orgs.Update))
	gorgs.POST("/:id/delete", orgs.Identified(orgs.Destroy))

	auth := &AuthHandler{
		Views: &HandlerViews{
			templates: template.Must(template.ParseGlob(dir + "/auths/*.html")),
		},
	}

	g := e.Group("/admin/auths")
	g.GET("", withFlash(auth.index))
	g.POST("", withFlash(auth.create))
	g.POST("/:id/disable", auth.Identified(auth.disable))
	g.POST("/:id/delete", auth.Identified(auth.destroy))

	return map[string]interface{}{
		"orgs":  orgs,
		"auths": auth,
	}
}
