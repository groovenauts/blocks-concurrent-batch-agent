//go:generate goagen bootstrap -d github.com/groovenauts/blocks-concurrent-batch-server/design

package server

import (
	"net/http"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/controller"
)

func init() {
	// Create service
	service := goa.New("appengine")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "InstanceGroup" controller
	c := controller.NewInstanceGroupController(service)
	app.MountInstanceGroupController(service, c)
	// Mount "InstanceGroupConstructionTask" controller
	c2 := controller.NewInstanceGroupConstructionTaskController(service)
	app.MountInstanceGroupConstructionTaskController(service, c2)
	// Mount "InstanceGroupDestructionTask" controller
	c3 := controller.NewInstanceGroupDestructionTaskController(service)
	app.MountInstanceGroupDestructionTaskController(service, c3)
	// Mount "InstanceGroupResizingTask" controller
	c4 := controller.NewInstanceGroupResizingTaskController(service)
	app.MountInstanceGroupResizingTaskController(service, c4)
	// Mount "Job" controller
	c5 := controller.NewJobController(service)
	app.MountJobController(service, c5)
	// Mount "Pipeline" controller
	c6 := controller.NewPipelineController(service)
	app.MountPipelineController(service, c6)
	// Mount "PipelineBase" controller
	c7 := controller.NewPipelineBaseController(service)
	app.MountPipelineBaseController(service, c7)
	// Mount "PipelineBaseClosingTask" controller
	c8 := controller.NewPipelineBaseClosingTaskController(service)
	app.MountPipelineBaseClosingTaskController(service, c8)
	// Mount "PipelineBaseOpeningTask" controller
	c9 := controller.NewPipelineBaseOpeningTaskController(service)
	app.MountPipelineBaseOpeningTaskController(service, c9)
	// Mount "dummy-auths" controller
	c10 := controller.NewDummyAuthsController(service)
	app.MountDummyAuthsController(service, c10)
	// Mount "swagger" controller
	c11 := controller.NewSwaggerController(service)
	app.MountSwaggerController(service, c11)

	// // Start service
	// if err := service.ListenAndServe(":8080"); err != nil {
	// 	service.LogError("startup", "err", err)
	// }

	http.HandleFunc("/", service.Mux.ServeHTTP)
}
