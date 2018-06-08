//go:generate goagen bootstrap -d github.com/groovenauts/blocks-concurrent-batch-server/design

package main

import (
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
)

func main() {
	// Create service
	service := goa.New("appengine")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "Auth" controller
	c := NewAuthController(service)
	app.MountAuthController(service, c)
	// Mount "InstanceGroupConstructingTask" controller
	c2 := NewInstanceGroupConstructingTaskController(service)
	app.MountInstanceGroupConstructingTaskController(service, c2)
	// Mount "InstanceGroupDestructingTask" controller
	c3 := NewInstanceGroupDestructingTaskController(service)
	app.MountInstanceGroupDestructingTaskController(service, c3)
	// Mount "InstanceGroupResizingTask" controller
	c4 := NewInstanceGroupResizingTaskController(service)
	app.MountInstanceGroupResizingTaskController(service, c4)
	// Mount "IntanceGroup" controller
	c5 := NewIntanceGroupController(service)
	app.MountIntanceGroupController(service, c5)
	// Mount "Job" controller
	c6 := NewJobController(service)
	app.MountJobController(service, c6)
	// Mount "Organization" controller
	c7 := NewOrganizationController(service)
	app.MountOrganizationController(service, c7)
	// Mount "Pipeline" controller
	c8 := NewPipelineController(service)
	app.MountPipelineController(service, c8)
	// Mount "PipelineBase" controller
	c9 := NewPipelineBaseController(service)
	app.MountPipelineBaseController(service, c9)
	// Mount "PipelineBaseClosingTask" controller
	c10 := NewPipelineBaseClosingTaskController(service)
	app.MountPipelineBaseClosingTaskController(service, c10)
	// Mount "PipelineBaseOpeningTask" controller
	c11 := NewPipelineBaseOpeningTaskController(service)
	app.MountPipelineBaseOpeningTaskController(service, c11)
	// Mount "swagger" controller
	c12 := NewSwaggerController(service)
	app.MountSwaggerController(service, c12)

	// Start service
	if err := service.ListenAndServe(":8080"); err != nil {
		service.LogError("startup", "err", err)
	}

}
