package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var JobMessage = Type("JobMessage", func() {
	Member("attributes", HashOf(String, String), "Attributes")
	Member("data", String, "Data")
})

var JobPayload = Type("JobPayload", func() {
	Member("message", JobMessage, "Job message")
	Member("id_by_client", String, "ID assigned by client")

	Required("message")
})

var Job = MediaType("application/vnd.job+json", func() {
	Description("job")
	Reference(JobPayload)

	attrNames := []string{
		"message",
		"id_by_client",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		Attribute("status", String, "Job Status", func() {
			Enum("inactive", "blocked", "publishing", "publishing_error",
				"published", "started", "success", "failure")
		})
		Attribute("pipeline_id", String, "Pipeline ID (UUID)")
		Attribute("pipeline_base_id", String, "PipelineBase ID (UUID)")
		Attribute("message_id", String, "Pubsub Message ID")
		Attribute("host_name", String, "Hostname where job is running")
		Attribute("publish_time", DateTime, "Time when job is published")
		Attribute("start_time", DateTime, "Time when job starts")
		Attribute("finish_time", DateTime, "Time when job finishes")
		UseTrait(TimestampsAttrTrait)

		requiredAttrs := append([]string{"id"}, attrNames...)
		Required(requiredAttrs...)
	})
	View("default", func() {
		Attribute("id")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		outputAttrs := []string{
			"status", "pipeline_id", "pipeline_base_id", "message_id",
			"host_name", // "output",
			"publish_time", "start_time", "finish_time",
		}
		for _, attrName := range outputAttrs {
			Attribute(attrName)
		}
		UseTrait(TimestampsViewTrait)
	})
})

var JobOutput = MediaType("application/vnd.job-output+json", func() {
	Description("job output")

	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("output", String, "Job output")
		Required("id", "output")
	})
	View("default", func() {
		Attribute("id")
		Attribute("output")
	})
})

var _ = Resource("Job", func() {
	BasePath("/jobs")
	DefaultMedia(Job)
	UseTrait(DefineResourceTrait)

	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("pipeline_id", String, "Pipeline ID")
			Param("pipeline_base_id", String, "Pipeline Base ID")
			Param("active", String, "Set true to activate soon")
		})
		Payload(JobPayload)
		Response(Created, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("output", func() {
		Description("output")
		Routing(GET("/:id/output"))
		Params(func() {
			Param("id")
		})
		Response(OK, JobOutput)
		UseTrait(DefaultResponseTrait)
	})
	Action("activate", func() {
		Description("Activate job")
		Routing(PUT("/:id/activate"))
		Params(func() {
			Param("id")
		})
		Response(OK, Job)
		Response(Created, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("inactivate", func() {
		Description("Inactivate job")
		Routing(PUT("/:id/inactivate"))
		Params(func() {
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("publishing_task", func() {
		Description("Publishing job task")
		Routing(PUT("/:id/publishing_task"))
		Params(func() {
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
})
