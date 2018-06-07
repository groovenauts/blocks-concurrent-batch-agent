package design

import (
	. "github.com/goadesign/goa/design/apidsl"
)

var _ = Resource("Constructing", func() {
	BasePath("/constructing_tasks")
	UseTrait(OperationResourceTrait)
})

var _ = Resource("Destructing", func() {
	BasePath("/destructing_tasks")
	UseTrait(OperationResourceTrait)
})

var _ = Resource("Resizing", func() {
	BasePath("/resizing_tasks")
	UseTrait(OperationResourceTrait)
})
