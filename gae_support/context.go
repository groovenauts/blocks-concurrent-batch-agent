package gae_support

import (
	"github.com/labstack/echo"

	"google.golang.org/appengine"
)

func With(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		req := c.Request()
		ctx := appengine.NewContext(req)
		c.Set("aecontext", ctx)
		return impl(c)
	}
}
