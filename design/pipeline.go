package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PipelinePayload = Type("PipelinePayload", func() {
	Member("name", String, "Name of pipeline_base", func() {
		Example("pipeline1")
	})
	Required("name")

	Reference(PipelineBasePayloadBody)
	members := []string{
		"instance_group",
		"container",
		"hibernation_delay",
	}
	for _, m := range members {
		Member(m)
	}
	Required(pipelineBasePayloadBodyRequired...)
})

var Pipeline = MediaType("application/vnd.pipeline+json", func() {
	Description("pipeline")
	Reference(PipelinePayload)

	inheritedAttrs := []string{
		"name",
		"instance_group",
		"container",
		"hibernation_delay",
	}

	Attributes(func() {
		UseTrait(IdTrait)
		for _, attr := range inheritedAttrs {
			Attribute(attr)
		}
		Attribute("prev_base_id", String, "Previous pipeline base ID")
		Attribute("curr_base_id", String, "Current pipeline base ID")
		Attribute("next_base_id", String, "Next pipeline base ID")
		Attribute("status", String, "Pipeline Status", func() {
			Enum("current_preparing", "current_preparing_error",
				"running", "next_preparing",
				"stopping", "stopping_error", "stopped")
		})
		UseTrait(TimestampsAttrTrait)

		Required(inheritedAttrs...)
	})
	View("default", func() {
		Attribute("id")
		for _, attr := range inheritedAttrs {
			Attribute(attr)
		}
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
		UseTrait(DefaultResponseTrait)
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
		UseTrait(DefaultResponseTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, Pipeline)
		UseTrait(DefaultResponseTrait)
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
		UseTrait(DefaultResponseTrait)
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
		UseTrait(DefaultResponseTrait)
	})
	Action("stop", func() {
		Description("Stop pipeline")
		Routing(PUT("/:id/stop"))
		Params(func() {
			Param("id")
		})
		Response(OK, Pipeline)
		UseTrait(DefaultResponseTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, Pipeline)
		UseTrait(DefaultResponseTrait)
	})
})
