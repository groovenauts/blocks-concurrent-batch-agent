package api

import (
	"gae_support"

	"github.com/labstack/echo"
)

var e *echo.Echo

func SetupRoutes(echo *echo.Echo) map[string]interface{} {
	e = echo

	h := &PipelineHandler{}
	h.buildActions()

	g := e.Group("/orgs/:org_id/pipelines")
	g.GET("", h.Actions["index"])
	g.POST("", h.Actions["create"])
	g.GET("/subscriptions", h.Actions["subscriptions"])

	g = e.Group("/pipelines")
	g.GET("/:id", h.Actions["show"])
	// g.POST("/:id/build_task", h.Actions["build"])
	g.POST("/:id/build_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.pipelineTask("build"))))))
	g.PUT("/:id/close", h.Actions["close"])
	// g.POST("/:id/close_task", h.Actions["close_task"])
	g.POST("/:id/close_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.pipelineTask("close"))))))
	g.DELETE("/:id", h.Actions["destroy"])

	g.GET("/refresh", h.Actions["refresh"])
	g.POST("/:id/refresh_task", h.Actions["refresh_task"])

	pjh := &PipelineJobHandler{}
	pjActions := pjh.buildActions()

	g = e.Group("/pipelines/:pipeline_id/jobs")
	g.GET("", pjActions["index"])
	g.POST("", pjActions["create"])

	g = e.Group("/jobs")
	g.GET("/:id", pjActions["show"])

	return map[string]interface{}{
		"pipelines": h,
		"pipeline_jobs": pjh,
	}
}
