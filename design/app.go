package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

const DefineTrait = "DefineTrait"

const TimestampCreatedAt = "created_at"
const TimestampUpdatedAt = "updated_at"

const TimestampsAttrTrait = "TimestampsAttrTrait"
const TimestampsViewTrait = "TimestampsViewTrait"

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

	Trait(TimestampsAttrTrait, func() {
		Attribute(TimestampCreatedAt, DateTime, "Datetime created")
		Attribute(TimestampUpdatedAt, DateTime, "Datetime updated")
	})
	Trait(TimestampsViewTrait, func() {
		Attribute(TimestampCreatedAt)
		Attribute(TimestampUpdatedAt)
	})
})
var _ = Resource("swagger", func() {
	Origin("*", func() {
		Methods("GET")
	})
	Files("swagger.json", "../swagger/swagger.json")
	Files("swagger/*filepath", "../static/swagger/")
})
