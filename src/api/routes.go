package api

import (
	"github.com/labstack/echo"
)

var e *echo.Echo

func SetupRoutes(echo *echo.Echo) map[string]interface{} {
	e = echo

	h := &PipelineHandler{}

	g := e.Group("/orgs/:org_id/pipelines")
	g.GET("", h.collection("org_id", h.index))
	g.POST("", h.collection("org_id", h.create))
	g.GET("/subscriptions", h.collection("org_id", h.subscriptions))

	g = e.Group("/pipelines")
	g.GET("/:id", h.member("id", h.show))
	g.PUT("/:id/cancel", h.member("id", h.cancel))
	g.PUT("/:id/close", h.member("id", h.cancel))
	g.POST("/:id/close_task", h.member("id", h.closeTask))
	g.DELETE("/:id", h.member("id", h.destroy))

	g.POST("/:id/build_task", h.member("id", h.buildTask))
	g.POST("/:id/wait_building_task", h.member("id", h.waitBuildingTask))
	g.POST("/:id/publish_task", h.member("id", h.publishTask))
	g.POST("/:id/subscribe_task", h.member("id", h.subscribeTask))
	g.POST("/:id/wait_closing_task", h.member("id", h.waitClosingTask))

	g.POST("/:id/refresh", h.member("id", h.refresh))
	g.POST("/:id/refresh_task", h.member("id", h.refreshTask))

	jh := &JobHandler{}

	g = e.Group("/pipelines/:pipeline_id/jobs")
	g.GET("", jh.collection("pipeline_id", jh.index))
	g.POST("", jh.collection("pipeline_id", jh.create))

	g = e.Group("/jobs")
	g.GET("/:id", jh.member("id", h.show))
	g.POST("/:id/getready", jh.member("id", jh.getReady))
	g.POST("/:id/wait_task", jh.member("id", jh.WaitToPublishTask))
	g.POST("/:id/publish_task", jh.member("id", jh.PublishTask))

	return map[string]interface{}{
		"pipelines": h,
		"jobs":      jh,
	}
}
