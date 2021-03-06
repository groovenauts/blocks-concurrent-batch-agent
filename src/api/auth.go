package api

import (
	"context"
	"net/http"
	"regexp"

	"github.com/labstack/echo"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/models"
)

const (
	AUTH_HEADER = "Authorization"
)

func withAuth(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		req := c.Request()
		raw := req.Header.Get(AUTH_HEADER)
		if raw == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		}
		re := regexp.MustCompile(`\ABearer\s+`)
		token := re.ReplaceAllString(raw, "")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		}
		org := c.Get("organization").(*models.Organization)
		_, err := org.AuthAccessor().FindWithToken(ctx, token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
		}
		return impl(c)
	}
}
