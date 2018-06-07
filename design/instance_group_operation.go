package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("Constructing", func() {
	BasePath("/constructing_tasks")
	DefaultMedia(Operation)
	Action("start", func() {
		Description("Start refreshing")
		Routing(POST(""))
		Params(func() {
			Param("id", String, "InstanceGroup ID")
		})
		Payload(OperationPayload)
		Response(Created, Operation)
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
	DefaultMedia(Operation)
	Action("start", func() {
		Description("Start refreshing")
		Routing(POST(""))
		Params(func() {
			Param("id", String, "InstanceGroup ID")
		})
		Payload(OperationPayload)
		Response(Created, Operation)
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
	DefaultMedia(Operation)
	Action("start", func() {
		Description("Start refreshing")
		Routing(POST(""))
		Params(func() {
			Param("id", String, "InstanceGroup ID")
		})
		Payload(OperationPayload)
		Response(Created, Operation)
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
