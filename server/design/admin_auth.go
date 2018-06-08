package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var Auth = MediaType("application/vnd.auth+json", func() {
	Description("auth")
	Attributes(func() {
		UseTrait(IdTrait)
		UseTrait(TimestampsAttrTrait)
		Required("id")
	})
	View("default", func() {
		Attribute("id")
		UseTrait(TimestampsViewTrait)
	})
})

var AuthSecret = MediaType("application/vnd.auth-secret+json", func() {
	Description("auth")
	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("token", String, "Token")
		UseTrait(TimestampsAttrTrait)
		Required("id")
	})
	View("default", func() {
		Attribute("id")
		Attribute("token")
		UseTrait(TimestampsViewTrait)
	})
})

var _ = Resource("Auth", func() {
	BasePath("/admin/organizations/:id/auths")
	DefaultMedia(Auth)
	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Params(func() {
			Param("org_id")
		})
		Response(OK, CollectionOf(Auth))
		UseTrait(DefineTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("org_id")
		})
		Response(Created, AuthSecret)
		UseTrait(DefineTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:auth_id"))
		Params(func() {
			Param("org_id")
			Param("id")
		})
		Response(OK, Auth)
		UseTrait(DefineTrait)
	})
})
