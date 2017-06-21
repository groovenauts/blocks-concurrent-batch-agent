package gae_support

import (
	"github.com/labstack/echo"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func With(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		req := c.Request()
		ctx := appengine.NewContext(req)
		log.Debugf(ctx, "gae_support.With\n")
		c.Set("aecontext", ctx)
		return impl(c)
	}
}
