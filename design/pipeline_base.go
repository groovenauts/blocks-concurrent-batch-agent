package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PipelineContainer = Type("PipelineContainer", func() {
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

var pipelineBasePayloadBodyRequired = []string{
	"instance_group", "container",
}
var PipelineBasePayloadBody = Type("PipelineBasePayloadBody", func() {
	Member("instance_group", InstanceGroupBody, "Instance Group configuration")
	Member("container", PipelineContainer, "Container configuration")
	Member("hibernation_delay", Integer, "Hibernation delay in seconds since last job finsihed")

	Required(pipelineBasePayloadBodyRequired...)
})

var PipelineBasePayload = Type("PipelineBasePayload", func() {
	Member("name", String, "Name of pipeline_base", func() {
		Example("pipeline1-123")
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

var PipelineBase = MediaType("application/vnd.pipeline-base+json", func() {
	Description("pipeline-base")
	Reference(PipelineBasePayload)

	attrNames := []string{
		"name",
		"instance_group",
		"container",
		"hibernation_delay",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		Attribute("status", String, "Pipeline Base Status", func() {
			Enum("opening", "opening_error", "hibernating", "waking", "waking_error",
				"awake", "hibernation_checking", "hibernation_going", "hibernation_going_error",
				"closing", "closing_error", "closed")
		})
		Attribute("instance_group_id", String, "ID of instance group")
		UseTrait(TimestampsAttrTrait)

		requiredAttrs := append([]string{"id"}, attrNames...)
		Required(requiredAttrs...)
	})
	View("default", func() {
		Attribute("id")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		Attribute("status")
		Attribute("instance_group_id")
		UseTrait(TimestampsViewTrait)
	})
})

var _ = Resource("PipelineBase", func() {
	BasePath("/pipeline_bases")
	DefaultMedia(PipelineBase)
	UseTrait(DefineResourceTrait)

	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Required("org_id")
		})
		Response(OK, CollectionOf(PipelineBase))
		UseTrait(DefaultResponseTrait)
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
		UseTrait(DefaultResponseTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
	Action("wakeup_done_task", func() {
		Description("Task to accept wakeup(construction) done notification")
		Routing(PUT("/:id/wakeup_done_task"))
		Params(func() {
			Param("id")
			Param("operation_id")
			Param("error")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
	Action("pull_task", func() {
		Description("Task to pull progress messages")
		Routing(PUT("/:id/pull_task"))
		Params(func() {
			Param("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
	Action("hibernation_checking_task", func() {
		Description("Task to check if it starts hibernation")
		Routing(PUT("/:id/hibernation_checking_task"))
		Params(func() {
			Param("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
	Action("hibernation_done_task", func() {
		Description("Task to accept hibernation(destruction) done notification")
		Routing(PUT("/:id/hibernation_done_task"))
		Params(func() {
			Param("id")
			Param("operation_id")
			Param("error")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
	Action("close", func() {
		Description("Close")
		Routing(PUT("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, PipelineBase)
		UseTrait(DefaultResponseTrait)
	})
})

// For pubsub (topic + subscription) * (job + progress)
var _ = Resource("PipelineBaseOpeningTask", func() {
	BasePath("/opening_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})

var _ = Resource("PipelineBaseClosingTask", func() {
	BasePath("/closing_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})
