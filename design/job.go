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
		Attribute("hostname", String, "Hostname where job is running")
		Attribute("published_at", DateTime, "Time when job is published")
		Attribute("started_at", DateTime, "Time when job starts")
		Attribute("finished_at", DateTime, "Time when job finishes")
		UseTrait(TimestampsAttrTrait)

		requiredAttrs := append([]string{"id", "status", "pipeline_base_id"}, attrNames...)
		Required(requiredAttrs...)
	})
	View("default", func() {
		Attribute("id")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		outputAttrs := []string{
			"status", "pipeline_id", "pipeline_base_id", "message_id",
			"hostname", // "output",
			"published_at", "started_at", "finished_at",
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


var JobActions = func() {
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
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
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("output", func() {
		Description("output")
		Routing(GET("/:id/output"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("id")
		})
		Response(OK, JobOutput)
		UseTrait(DefaultResponseTrait)
	})
	Action("activate", func() {
		Description("Activate job")
		Routing(PUT("/:id/activate"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
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
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
}

var _ = Resource("PipelineJob", func() {
	BasePath("/orgs/:org_id/pipelines/:name/jobs")
	DefaultMedia(Job)
	UseTrait(DefineResourceTrait)

	JobActions()
})

var _ = Resource("PipelineBaseJob", func() {
	BasePath("/orgs/:org_id/pipeline_bases/:name/jobs")
	DefaultMedia(Job)
	UseTrait(DefineResourceTrait)

	// JobActions() // same as PipelineJob

	Action("publishing_task", func() {
		Description("Publishing job task")
		Routing(PUT("/:id/publishing_task"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("id")
		})
		Response(OK, Job)
		UseTrait(DefaultResponseTrait)
	})
})
