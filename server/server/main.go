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

	// Mount "Auth" controller
	c := controller.NewAuthController(service)
	app.MountAuthController(service, c)
	// Mount "InstanceGroupConstructingTask" controller
	c2 := controller.NewInstanceGroupConstructingTaskController(service)
	app.MountInstanceGroupConstructingTaskController(service, c2)
	// Mount "InstanceGroupDestructingTask" controller
	c3 := controller.NewInstanceGroupDestructingTaskController(service)
	app.MountInstanceGroupDestructingTaskController(service, c3)
	// Mount "InstanceGroupResizingTask" controller
	c4 := controller.NewInstanceGroupResizingTaskController(service)
	app.MountInstanceGroupResizingTaskController(service, c4)
	// Mount "IntanceGroup" controller
	c5 := controller.NewIntanceGroupController(service)
	app.MountIntanceGroupController(service, c5)
	// Mount "Job" controller
	c6 := controller.NewJobController(service)
	app.MountJobController(service, c6)
	// Mount "Organization" controller
	c7 := controller.NewOrganizationController(service)
	app.MountOrganizationController(service, c7)
	// Mount "Pipeline" controller
	c8 := controller.NewPipelineController(service)
	app.MountPipelineController(service, c8)
	// Mount "PipelineBase" controller
	c9 := controller.NewPipelineBaseController(service)
	app.MountPipelineBaseController(service, c9)
	// Mount "PipelineBaseClosingTask" controller
	c10 := controller.NewPipelineBaseClosingTaskController(service)
	app.MountPipelineBaseClosingTaskController(service, c10)
	// Mount "PipelineBaseOpeningTask" controller
	c11 := controller.NewPipelineBaseOpeningTaskController(service)
	app.MountPipelineBaseOpeningTaskController(service, c11)
	// Mount "swagger" controller
	c12 := controller.NewSwaggerController(service)
	app.MountSwaggerController(service, c12)

	// // Start service
	// if err := service.ListenAndServe(":8080"); err != nil {
	// 	service.LogError("startup", "err", err)
	// }

	http.HandleFunc("/", service.Mux.ServeHTTP)
}
