package controller

import (
	"context"
	"net/http"
	"regexp"

	"google.golang.org/appengine"

	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
	"github.com/goadesign/goa"
)

// See https://github.com/goadesign/examples/blob/master/security/api_key.go

var BearerPattern = regexp.MustCompile(`\ABearer\s+`)

var (
	// ErrUnauthorized is the error returned for unauthorized requests.
	ErrUnauthorized = goa.NewErrorClass("unauthorized", 401)
)

// NewAPIKeyMiddleware creates a middleware that checks for the presence of an authorization header
// and validates its content.
func NewAuthorizationMiddleware() goa.Middleware {
	// Instantiate API Key security scheme details generated from design
	scheme := app.NewAPIKeySecurity()

	// Middleware
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			// Retrieve and log header specified by scheme
			key := req.Header.Get(scheme.Name)
			// A real app would do something more interesting here
			if len(key) == 0 || key == "Bearer" {
				goa.LogInfo(ctx, "failed api key auth")
				return ErrUnauthorized("missing auth")
			}
			token := BearerPattern.ReplaceAllString(key, "")
			if token == "" {
				return ErrUnauthorized("missing auth token")
			}

			// org := c.Get("organization").(*models.Organization)
			// _, err := org.AuthAccessor().FindWithToken(ctx, token)
			acc := &model.AuthAccessor{}
			appCtx := appengine.NewContext(req)
			_, err := acc.FindWithToken(appCtx, token)
			if err != nil {
				return ErrUnauthorized("invalid token")
			}

			// Proceed.
			goa.LogInfo(ctx, "auth", "apikey", "key", key)
			return h(ctx, rw, req)
		}
	}
}
