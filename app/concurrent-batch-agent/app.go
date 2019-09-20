package main

import (
	"net/http"

	"github.com/labstack/echo"

	"admin"
	"api"
)

func main() {
	e := echo.New()
	// note: we don't need to provide the middleware or static handlers, that's taken care of by the platform
	// app engine has it's own "main" wrapper - we just need to hook echo into the default handler

	admin.SetupRoutes(e, "admin/views")
	api.SetupRoutes(e)

	http.Handle("/", e)
}

// reference our echo instance and create it early
