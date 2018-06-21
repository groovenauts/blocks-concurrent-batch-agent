package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var Operation = MediaType("application/vnd.instance-group-operation+json", func() {
	Description("instance-group-operation")
	attrNames := []string{
		"id",
		"owner_type",
		"owner_id",
		"name",
		"service",
		"operation_type",
		"status",
		"project_id",
		"zone",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("owner_type", String, "Owner type name", func() {
			Example("InstanceGroup")
		})
		Attribute("owner_id", String, "Owner id", func() {
			Example("bd2d5ee3-d8be-4024-85a7-334dee9c1c88")
		})

		Attribute("name", String, "Name", func() {
			Example("instancegroup-ope1")
		})
		Attribute("service", String, "Service name", func() {
			Example("deploymentmanager")
		})
		Attribute("operation_type", String, "Operation Type")
		Attribute("status", String, "Operation Status")

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
