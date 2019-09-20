package main

import (
	"net/http"

	"google.golang.org/appengine"

	"github.com/labstack/echo"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/admin"
	"github.com/groovenauts/blocks-concurrent-batch-server/src/api"
)

func main() {
	e := echo.New()
	// note: we don't need to provide the middleware or static handlers, that's taken care of by the platform
	// app engine has it's own "main" wrapper - we just need to hook echo into the default handler

	admin.SetupRoutes(e, "admin/views")
	api.SetupRoutes(e)

	http.Handle("/", e)
	appengine.Main()
}

// reference our echo instance and create it early
