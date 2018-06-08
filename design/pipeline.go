package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PipelinePayload = Type("PipelinePayload", func() {
	Member("base", PipelineBasePayload, "PipelineBase configuration")

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
