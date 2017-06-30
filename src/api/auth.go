package api

import (
	"net/http"
	"regexp"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
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

func withPlIDHexAuth(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
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
		pl := c.Get("pipeline").(*models.Pipeline)
		if pl.IDHex() != token {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
		}
		return impl(c)
	}

}
