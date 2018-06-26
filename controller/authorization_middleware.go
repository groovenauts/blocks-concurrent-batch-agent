package controller

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// See https://github.com/goadesign/examples/blob/master/security/api_key.go

var BearerPattern = regexp.MustCompile(`\ABearer\s+`)

var (
	// ErrUnauthorized is the error returned for unauthorized requests.
	ErrUnauthorized = goa.NewErrorClass("unauthorized", 401)
)

const ContextOrgKey = "organization.key"

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
			auth, err := acc.FindWithToken(appCtx, token)
			if err != nil {
				return ErrUnauthorized("invalid token")
			}

			orgKey, err := auth.OrganizationKey()
			if err != nil {
				return ErrUnauthorized(fmt.Sprintf("OrganizationKey error because of %v", err))
			}
			if orgKey == nil {
				return ErrUnauthorized("Organization Key not found")
			}
			ctx = context.WithValue(ctx, ContextOrgKey, orgKey)

			// Proceed.
			goa.LogInfo(ctx, "auth", "apikey", "key", key)
			return h(ctx, rw, req)
		}
	}
}

func WithAuthOrgKey(ctx context.Context, f func(*datastore.Key) error) error {
	obj := ctx.Value(ContextOrgKey)
	if obj == nil {
		return ErrUnauthorized(fmt.Sprintf("%s not found in context", ContextOrgKey))
	}
	orgKey, ok := obj.(*datastore.Key)
	if !ok {
		return ErrUnauthorized(fmt.Sprintf("%s in context isn't a *datastore.Key", ContextOrgKey))
	}
	return f(orgKey)
}
