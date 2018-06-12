package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var OperationPayload = Type("OperationPayload", func() {
	Member("owner_type", String, "Owner type name", func() {
		Example("InstanceGroup")
	})
	Member("owner_id", String, "Owner id", func() {
		Example("bd2d5ee3-d8be-4024-85a7-334dee9c1c88")
	})
	Member("finalize_task_path", String, "Callback task path")

	Member("name", String, "Name", func() {
		Example("instancegroup-ope1")
	})
	Member("service", String, "Service name", func() {
		Example("deploymentmanager")
	})
	Member("operation_type", String, "Operation Type")
	Member("status", String, "Operation Status")

	Member("project_id", String, "GCP Project ID", func() {
		Example("dummy-proj-999")
	})
	Member("zone", String, "GCP zone", func() {
		Example("us-central1-f")
	})

	Required("name", "service", "operation_type", "status", "project_id", "zone")
})

var Operation = MediaType("application/vnd.instance-group-operation+json", func() {
	Description("instance-group-operation")
	Reference(OperationPayload)
	attrNames := []string{
		"name",
		"service",
		"operation_type",
		"status",
		"project_id",
		"zone",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
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
