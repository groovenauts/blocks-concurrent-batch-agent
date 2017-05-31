package main

import (
	"github.com/groovenauts/blocks-concurrent-batch-agent/admin"
)

func init() {
	admin.Setup(e, "admin/views")
}
