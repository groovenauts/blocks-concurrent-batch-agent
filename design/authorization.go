package design

import (
	. "github.com/goadesign/goa/design/apidsl"
)

// APIKey defines a security scheme using an API key (shared secret).  The scheme uses the
// "Authorization" header to lookup the key.
var Authorization = APIKeySecurity("api_key", func() {
	Description("Shared password")
	Header("Authorization")
})
