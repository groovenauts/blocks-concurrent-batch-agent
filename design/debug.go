package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)


var DummyAuth = MediaType("application/vnd.dummy-auth+json", func() {
	Description("Dummy auth")

	Attributes(func() {
		Attribute("organization_id", String)
		Attribute("token", String)
		Required("organization_id", "token")
	})
	View("default", func() {
		Attribute("organization_id")
		Attribute("token")
	})
})

var _ = Resource("dummy-auths", func() {
	BasePath("/dummy-auths")
	DefaultMedia(DummyAuth)
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Response(Created, DummyAuth)
		UseTrait(DefineTrait)
	})
})
