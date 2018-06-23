package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

const DefaultResponseTrait = "DefaultResponseTrait"
const DefineResourceTrait = "DefineResourceTrait"

const IdTrait = "IdTrait"

const TimestampCreatedAt = "created_at"
const TimestampUpdatedAt = "updated_at"

const TimestampsAttrTrait = "TimestampsAttrTrait"
const TimestampsViewTrait = "TimestampsViewTrait"

const TaskResponsesTrait = "TaskResponsesTrait"

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
	Trait(DefaultResponseTrait, func() {
		Response(Unauthorized, ErrorMedia)
		Response(NotFound, ErrorMedia)
		Response(BadRequest, ErrorMedia)
		Response(InternalServerError, ErrorMedia)
	})
	Trait(DefineResourceTrait, func() {
		Security(Authorization)
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

	Trait(TaskResponsesTrait, func() {
		Response(OK, CloudAsyncOperation)           // 200 (他のなにかによって)既に完了済み
		Response(Created, CloudAsyncOperation)      // 201 継続
		Response(Accepted, CloudAsyncOperation)     // 202 完了
		Response(NoContent, CloudAsyncOperation)    // 204 エラー
		Response(ResetContent, CloudAsyncOperation) // 205 成功(再実行)
		UseTrait(DefaultResponseTrait)
	})

	Trait(CloudAsyncOperationResourceTrait, CloudAsyncOperationResourceTraitFunc)
})
var _ = Resource("swagger", func() {
	Origin("*", func() {
		Methods("GET")
	})
	Files("swagger.json", "../swagger/swagger.json")
	Files("swagger/*filepath", "../static/swagger/")
})
