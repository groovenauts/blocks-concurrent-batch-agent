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
	h := &PipelineHandler{}

	cActions := h.buildCollectionActions("org_id")
	g := e.Group("/orgs/:org_id/pipelines")
	g.GET("", cActions["index"])
	g.POST("", cActions["create"])
	g.GET("/subscriptions", cActions["subscriptions"])

	mActions := h.buildMemberActions("id")
	g = e.Group("/pipelines")
	g.GET("/:id", mActions["show"])
	g.PUT("/:id/cancel", mActions["cancel"])
	g.PUT("/:id/close", mActions["cancel"])
	g.POST("/:id/close_task", mActions["closeTask"])
	g.DELETE("/:id", mActions["destroy"])

	g.POST("/:id/build_task", mActions["buildTask"])
	g.POST("/:id/wait_building_task", mActions["waitBuildingTask"])
	g.POST("/:id/publish_task", mActions["publishTask"])
	g.POST("/:id/subscribe_task", mActions["subscribeTask"])
	g.POST("/:id/wait_closing_task", mActions["waitClosingTask"])
	g.POST("/:id/refresh", mActions["refresh"])
	g.POST("/:id/refresh_task", mActions["refreshTask"])

	return h
}

func SetupRoutesOfJobs(e *echo.Echo) *JobHandler {
	jh := &JobHandler{}

	cActions := jh.buildCollectionActions("pipeline_id")
	g := e.Group("/pipelines/:pipeline_id/jobs")
	g.GET("", cActions["index"])
	g.POST("", cActions["create"])

	mActions := jh.buildMemberActions("id")
	g = e.Group("/jobs")
	g.GET("/:id", mActions["show"])
	g.POST("/:id/getready", mActions["getready"])
	g.POST("/:id/wait_task", mActions["wait_task"])
	g.POST("/:id/publish_task", mActions["publish_task"])

	return jh
}
