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

	// Mount "InstanceGroupConstructingTask" controller
	c := NewInstanceGroupConstructingTaskController(service)
	app.MountInstanceGroupConstructingTaskController(service, c)
	// Mount "InstanceGroupDestructingTask" controller
	c2 := NewInstanceGroupDestructingTaskController(service)
	app.MountInstanceGroupDestructingTaskController(service, c2)
	// Mount "InstanceGroupResizingTask" controller
	c3 := NewInstanceGroupResizingTaskController(service)
	app.MountInstanceGroupResizingTaskController(service, c3)
	// Mount "IntanceGroup" controller
	c4 := NewIntanceGroupController(service)
	app.MountIntanceGroupController(service, c4)
	// Mount "Job" controller
	c5 := NewJobController(service)
	app.MountJobController(service, c5)
	// Mount "Pipeline" controller
	c6 := NewPipelineController(service)
	app.MountPipelineController(service, c6)
	// Mount "PipelineBase" controller
	c7 := NewPipelineBaseController(service)
	app.MountPipelineBaseController(service, c7)
	// Mount "PipelineBaseClosingTask" controller
	c8 := NewPipelineBaseClosingTaskController(service)
	app.MountPipelineBaseClosingTaskController(service, c8)
	// Mount "PipelineBaseOpeningTask" controller
	c9 := NewPipelineBaseOpeningTaskController(service)
	app.MountPipelineBaseOpeningTaskController(service, c9)
	// Mount "swagger" controller
	c10 := NewSwaggerController(service)
	app.MountSwaggerController(service, c10)

	// Start service
	if err := service.ListenAndServe(":8080"); err != nil {
		service.LogError("startup", "err", err)
	}

}
