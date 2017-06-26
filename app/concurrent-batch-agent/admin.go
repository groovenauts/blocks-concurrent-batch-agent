package main

import (
	"admin"
)

func init() {
	admin.SetupRoutes(e, "admin/views")
}
