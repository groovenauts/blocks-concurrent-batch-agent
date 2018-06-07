package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var OrganizationPayload = Type("OrganizationPayload", func() {
	Member("name", String, "Name", func() {
		Example("org1")
	})
	Member("memo", String, "Memo")
	Member("token_amount", Integer, "Token Amount", func() {
		Default(100)
		Example(10)
	})
})

var Organization = MediaType("application/vnd.organization+json", func() {
	Description("organization")
	Reference(OrganizationPayload)
	attrNames := []string{
		"name",
		"memo",
		"token_amount",
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

var _ = Resource("Organization", func() {
	BasePath("/admin/organizations")
	DefaultMedia(Organization)
	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Response(OK, CollectionOf(Organization))
		UseTrait(DefineTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Payload(OrganizationPayload)
		Response(Created, Organization)
		UseTrait(DefineTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, Organization)
		UseTrait(DefineTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, Organization)
		UseTrait(DefineTrait)
	})
})
