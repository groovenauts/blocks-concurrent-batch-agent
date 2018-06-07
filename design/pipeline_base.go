package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PipelineContainerPayload = Type("PipelineContainerPayload", func() {
	Member("name", String, "Container name")
	Member("size", Integer, "Container size per VM", func() {
		Default(1)
		Example(2)
	})
	Member("command", String, "Command for docker run", func() {
		Example("bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}")
	})
	Member("options", String, "Options for docker run", func() {
		Example("--restart=on-failure:3")
	})
	Member("stackdriver_agent", Boolean, "Use stackdriver agent")
	Required("name")
})

var PipelineBasePayload = Type("PipelineBasePayload", func() {
	Member("instance_group", InstanceGroupPayload, "Instance Group configuration")
	Member("container", PipelineContainerPayload, "Container configuration")
	Member("hibernation_delay", Integer, "Hibernation delay in seconds since last job finsihed")

	Required("instance_group", "container")
})

var PipelineBase = MediaType("application/vnd.pipeline-base+json", func() {
	Description("pipeline-base")
	Reference(PipelineBasePayload)

	attrNames := []string{
		"instance_group",
		"container",
		"hibernation_delay",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		Attribute("status", String, "Status", func() {
			Enum("opening", "opening_error", "hibernating", "waking", "waking_error",
				"awake", "hibernation_checking", "hibernation_going", "hibernation_going_error",
				"closing", "closing_error", "closed")
		})
		UseTrait(TimestampsAttrTrait)

		requiredAttrs := append([]string{"id"}, attrNames...)
		Required(requiredAttrs...)
	})
	View("default", func() {
		Attribute("id")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		UseTrait(TimestampsViewTrait)
	})
})

var _ = Resource("PipelineBase", func() {
	BasePath("/pipeline_bases")
	DefaultMedia(PipelineBase)
	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Required("org_id")
		})
		Response(OK, CollectionOf(PipelineBase))
		UseTrait(DefineTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Required("org_id")
		})
		Payload(PipelineBasePayload)
		Response(Created, PipelineBase)
		UseTrait(DefineTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefineTrait)
	})
	Action("close", func() {
		Description("Close")
		Routing(PUT("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefineTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefineTrait)
	})
})
