package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var CloudAsyncOperation = MediaType("application/vnd.cloud-async-operation+json", func() {
	Description("CloudAsyncOperation")
	attrNames := []string{
		"id",
		"owner_type",
		"owner_name",
		"name",
		"service",
		"operation_type",
		"status",
		"project_id",
		"zone",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("name", String, "Name", func() {
			Example("instancegroup-ope1")
		})
		Attribute("service", String, "Service name", func() {
			Example("deploymentmanager")
			Example("custom") // For health check
		})
		Attribute("operation_type", String, "Operation Type", func() {
			Example("insert")
			Example("update")
			Example("delete")
		})
		Attribute("status", String, "Operation Status", func() {
			Example("PENDING")
			Example("RUNNING")
			Example("DONE")
		})

		Attribute("project_id", String, "GCP Project ID", func() {
			Example("dummy-proj-999")
		})
		Attribute("zone", String, "GCP zone", func() {
			Example("us-central1-f")
		})
		UseTrait(TimestampsAttrTrait)

		Required(attrNames...)
	})
	View("default", func() {
		Attribute("id")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		UseTrait(TimestampsViewTrait)
	})
})

const CloudAsyncOperationResourceTrait = "CloudAsyncOperationResourceTrait"

func CloudAsyncOperationResourceTraitFunc() {
	DefaultMedia(CloudAsyncOperation)
	Action("start", func() {
		Description("Start operation")
		Routing(POST(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name", String, "Resource Name")
		})
		UseTrait(TaskResponsesTrait)
	})

	Action("watch", func() {
		Description("Watch")
		Routing(PUT("/:id"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name", String, "Resource Name")
			Param("id", String)
		})
		UseTrait(TaskResponsesTrait)
	})
}
