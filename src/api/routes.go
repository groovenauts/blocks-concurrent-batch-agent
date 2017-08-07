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
	g.PUT("/:id/close", h.Actions["close"])
	g.POST("/:id/close_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.closeTask)))))
	g.DELETE("/:id", h.Actions["destroy"])

	g.POST("/:id/build_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.buildTask)))))
	g.POST("/:id/wait_building_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.waitBuildingTask)))))
	g.POST("/:id/publish_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.publishTask)))))
	g.POST("/:id/subscribe_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.subscribeTask)))))
	g.POST("/:id/start_closing_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.startClosingTask)))))
	g.POST("/:id/wait_closing_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.waitClosingTask)))))

	g.POST("/:id/refresh", h.Actions["refresh"])
	g.POST("/:id/refresh_task", gae_support.With(plBy("id", PlToOrg(withAuth(h.refreshTask)))))

	jh := &JobHandler{}
	jhActions := jh.buildActions()

	g = e.Group("/pipelines/:pipeline_id/jobs")
	g.GET("", jhActions["index"])
	g.POST("", jhActions["create"])

	g = e.Group("/jobs")
	g.GET("/:id", jhActions["show"])
	g.POST("/:id/getready", jhActions["getready"])
	g.POST("/:id/wait_task", gae_support.With(jobBy("id", JobToPl(PlToOrg(withAuth(jh.WaitToPublishTask))))))
	g.POST("/:id/publish_task", gae_support.With(jobBy("id", JobToPl(PlToOrg(withAuth(jh.PublishTask))))))

	return map[string]interface{}{
		"pipelines": h,
		"jobs":      jh,
	}
}
