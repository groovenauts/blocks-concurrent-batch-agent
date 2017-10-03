package api

import (
	"github.com/labstack/echo"
)

var e *echo.Echo

func SetupRoutes(echo *echo.Echo) map[string]interface{} {
	e = echo // e is the global variable defined aborve.
	return map[string]interface{}{
		"pipelines": SetupRoutesOfPipelines(echo),
		"jobs":      SetupRoutesOfJobs(echo),
	}
}

func SetupRoutesOfPipelines(e *echo.Echo) *PipelineHandler {
	h := &PipelineHandler{
		org_id_name:      "org_id",
		pipeline_id_name: "id",
	}

	g := e.Group("/orgs/:org_id/pipelines")
	g.GET("", h.collection(h.index))
	g.POST("", h.collection(h.create))
	g.GET("/subscriptions", h.collection(h.subscriptions))

	g = e.Group("/pipelines")
	g.GET("/:id", h.member(h.show))
	g.PUT("/:id/cancel", h.member(h.cancel))
	g.PUT("/:id/close", h.member(h.cancel))
	g.POST("/:id/close_task", h.member(h.closeTask))
	g.DELETE("/:id", h.member(h.destroy))

	g.POST("/:id/build_task", h.member(h.buildTask))
	g.POST("/:id/wait_building_task", h.member(h.waitBuildingTask))
	g.POST("/:id/publish_task", h.member(h.publishTask))
	g.POST("/:id/subscribe_task", h.member(h.subscribeTask))
	g.POST("/:id/wait_closing_task", h.member(h.waitClosingTask))
	g.POST("/:id/refresh", h.member(h.refresh))
	g.POST("/:id/refresh_task", h.member(h.refreshTask))

	return h
}

func SetupRoutesOfJobs(e *echo.Echo) *JobHandler {
	h := &JobHandler{
		pipeline_id_name: "pipeline_id",
		job_id_name:      "id",
	}

	g := e.Group("/pipelines/:pipeline_id/jobs")
	g.GET("", h.collection(h.index))
	g.POST("", h.collection(h.create))

	g = e.Group("/jobs")
	g.GET("/:id", h.member(h.show))
	g.POST("/:id/getready", h.member(h.getReady))
	g.POST("/:id/wait_task", h.member(h.WaitToPublishTask))
	g.POST("/:id/publish_task", h.member(h.PublishTask))

	return h
}
