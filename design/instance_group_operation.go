package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var InstanceGroupOperationPayload = Type("InstanceGroupOperationPayload", func() {
	Member("name", String, "Name", func() {
		Example("instancegroup-ope1")
	})
	Member("service", String, "Service name", func() {
		Example("deploymentmanager")
	})
	Member("operation_type", String, "Operation Type")
	Member("status", String, "status")

	Member("project_id", String, "GCP Project ID", func() {
		Example("dummy-proj-999")
	})
	Member("zone", String, "GCP zone", func() {
		Example("us-central1-f")
	})

	Required("name", "service", "operation_type", "status", "project_id", "zone")
})

var InstanceGroupOperation = MediaType("application/vnd.instance-group-operation+json", func() {
	Description("instance-group-operation")
	Reference(InstanceGroupOperationPayload)
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

var _ = Resource("Constructing", func() {
	BasePath("/constructing_tasks")
	DefaultMedia(InstanceGroupOperation)
	Action("start", func() {
		Description("Start refreshing")
		Routing(POST(""))
		Params(func() {
			Param("id", String, "InstanceGroup ID")
		})
		Payload(InstanceGroupOperationPayload)
		Response(Created, InstanceGroupOperation)
		UseTrait(DefineTrait)
	})
	Action("refresh", func() {
		Description("Refresh")
		Routing(PUT("/:id"))
		Params(func() {
			Param("id")
		})
		Response(Accepted, InstanceGroup)
		Response(OK, InstanceGroup)
		UseTrait(DefineTrait)
	})
})

var _ = Resource("Destructing", func() {
	BasePath("/destructing_tasks")
	DefaultMedia(InstanceGroupOperation)
	Action("start", func() {
		Description("Start refreshing")
		Routing(POST(""))
		Params(func() {
			Param("id", String, "InstanceGroup ID")
		})
		Payload(InstanceGroupOperationPayload)
		Response(Created, InstanceGroupOperation)
		UseTrait(DefineTrait)
	})
	Action("refresh", func() {
		Description("Refresh")
		Routing(PUT("/:id"))
		Params(func() {
			Param("id")
		})
		Response(Accepted, InstanceGroup)
		Response(OK, InstanceGroup)
		UseTrait(DefineTrait)
	})
})

var _ = Resource("Resizing", func() {
	BasePath("/resizing_tasks")
	DefaultMedia(InstanceGroupOperation)
	Action("start", func() {
		Description("Start refreshing")
		Routing(POST(""))
		Params(func() {
			Param("id", String, "InstanceGroup ID")
		})
		Payload(InstanceGroupOperationPayload)
		Response(Created, InstanceGroupOperation)
		UseTrait(DefineTrait)
	})
	Action("refresh", func() {
		Description("Refresh")
		Routing(PUT("/:id"))
		Params(func() {
			Param("id")
		})
		Response(Accepted, InstanceGroup)
		Response(OK, InstanceGroup)
		UseTrait(DefineTrait)
	})
})
