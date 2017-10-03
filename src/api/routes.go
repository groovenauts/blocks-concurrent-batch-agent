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

	g := e.Group("/orgs/:org_id/pipelines", h.collection)
	g.GET("", h.index)
	g.POST("", h.create)
	g.GET("/subscriptions", h.subscriptions)

	g = e.Group("/pipelines", h.member)
	g.GET("/:id", h.show)
	g.PUT("/:id/cancel", h.cancel)
	g.PUT("/:id/close", h.cancel)
	g.POST("/:id/close_task", h.closeTask)
	g.DELETE("/:id", h.destroy)

	g.POST("/:id/build_task", h.buildTask)
	g.POST("/:id/wait_building_task", h.waitBuildingTask)
	g.POST("/:id/publish_task", h.publishTask)
	g.POST("/:id/subscribe_task", h.subscribeTask)
	g.POST("/:id/wait_closing_task", h.waitClosingTask)
	g.POST("/:id/refresh", h.refresh)
	g.POST("/:id/refresh_task", h.refreshTask)

	return h
}

func SetupRoutesOfJobs(e *echo.Echo) *JobHandler {
	h := &JobHandler{
		pipeline_id_name: "pipeline_id",
		job_id_name:      "id",
	}

	g := e.Group("/pipelines/:pipeline_id/jobs", h.collection)
	g.GET("", h.index)
	g.POST("", h.create)

	g = e.Group("/jobs", h.member)
	g.GET("/:id", h.show)
	g.POST("/:id/getready", h.getReady)
	g.POST("/:id/wait_task", h.WaitToPublishTask)
	g.POST("/:id/publish_task", h.PublishTask)

	return h
}
