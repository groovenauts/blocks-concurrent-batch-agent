package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

const DefineTrait = "DefineTrait"

const IdTrait = "IdTrait"

const TimestampCreatedAt = "created_at"
const TimestampUpdatedAt = "updated_at"

const TimestampsAttrTrait = "TimestampsAttrTrait"
const TimestampsViewTrait = "TimestampsViewTrait"

const OperationResourceTrait = "OperationResourceTrait"

var _ = API("appengine", func() {
	Title("The appengine example")
	Description("A simple appengine example")
	Host("localhost:8080")
	Scheme("http")
	BasePath("/")
	Origin("*", func() {
		Methods("GET", "POST", "PUT", "DELETE", "OPTIONS")
		MaxAge(600)
		Credentials()
	})
	Trait(DefineTrait, func() {
		Response(Unauthorized, ErrorMedia)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})

	Trait(IdTrait, func() {
		Attribute("id", String, "ID", func() {
			Example("bd2d5ee3-d8be-4024-85a7-334dee9c1c88")
		})
	})
	Trait(TimestampsAttrTrait, func() {
		Attribute(TimestampCreatedAt, DateTime, "Datetime created")
		Attribute(TimestampUpdatedAt, DateTime, "Datetime updated")
	})
	Trait(TimestampsViewTrait, func() {
		Attribute(TimestampCreatedAt)
		Attribute(TimestampUpdatedAt)
	})

	Trait(OperationResourceTrait, func() {
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
})
var _ = Resource("swagger", func() {
	Origin("*", func() {
		Methods("GET")
	})
	Files("swagger.json", "../swagger/swagger.json")
	Files("swagger/*filepath", "../static/swagger/")
})
