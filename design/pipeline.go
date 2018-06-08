package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PipelinePayload = Type("PipelinePayload", func() {
	Member("name", String, "Name of pipeline_base", func(){
		Example("pipeline1")
	})
	Required("name")

	Member("base", PipelineBasePayloadBody, "PipelineBase configuration")
	Required("base")
})

var Pipeline = MediaType("application/vnd.pipeline+json", func() {
	Description("pipeline")
	Reference(PipelinePayload)

	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("base")
		Attribute("prev_base_id", String, "Previous pipeline base ID")
		Attribute("curr_base_id", String, "Current pipeline base ID")
		Attribute("next_base_id", String, "Next pipeline base ID")
		Attribute("status", String, "Status", func() {
			Enum("current_preparing", "current_preparing_error",
				"running", "next_preparing",
				"stopping", "stopping_error", "stopped")
		})
		UseTrait(TimestampsAttrTrait)

		Required("base", "status")
	})
	View("default", func() {
		Attribute("id")
		Attribute("base")
		Attribute("prev_base_id")
		Attribute("curr_base_id")
		Attribute("next_base_id")
		Attribute("status")
		UseTrait(TimestampsViewTrait)
	})
})

var _ = Resource("Pipeline", func() {
	BasePath("/pipelines")
	DefaultMedia(Pipeline)
	UseTrait(DefineResourceTrait)

	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Required("org_id")
		})
		Response(OK, CollectionOf(Pipeline))
		UseTrait(DefineTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Required("org_id")
		})
		Payload(PipelinePayload)
		Response(Created, Pipeline)
		UseTrait(DefineTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, Pipeline)
		UseTrait(DefineTrait)
	})
	Action("preparing_finalize_task", func() {
		Description("Task to finalize current_preparing or next_preparing status")
		Routing(PUT("/:id/preparing_finalize_task"))
		Params(func() {
			Param("id")
			Param("operation_id")
			Param("error")
		})
		Response(OK, Pipeline)
		UseTrait(DefineTrait)
	})
	Action("current", func() {
		Description("Update current pipeline base")
		Routing(PUT("/:id/current"))
		Params(func() {
			Param("id")
			Param("pipeline_base_id")
			Required("pipeline_base_id")
		})
		Response(OK, Pipeline)
		UseTrait(DefineTrait)
	})
	Action("stop", func() {
		Description("Stop pipeline")
		Routing(PUT("/:id/stop"))
		Params(func() {
			Param("id")
		})
		Response(OK, Pipeline)
		UseTrait(DefineTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, Pipeline)
		UseTrait(DefineTrait)
	})
})
