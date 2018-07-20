package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var InstanceGroupVmDisk = Type("InstanceGroupVmDisk", func() {
	// Member("disk_name", String, "Disk name")
	Member("disk_size_gb", Integer, "Disk size", func() {
		Example(50)
	})
	Member("disk_type", String, "Disk type", func() {
		Example("projects/dummy-proj-999/zones/asia-east1-a/diskTypes/pd-standard")
	})
	Member("source_image", String, "Source image", func() {
		Example("https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable")
	})
	Required("source_image")
})

var InstanceGroupAccelerators = Type("InstanceGroupAccelerators", func() {
	Member("count", Integer, "Count", func() {
		Example(2)
	})
	Member("type", String, "Type", func() {
		Example("nvidia-tesla-p100")
	})
	Required("count", "type")
})

var InstanceGroupHealthCheckConfig = Type("InstanceGroupHealthCheckConfig", func() {
	Member("interval", Integer, "Interval in second", func() {
		Default(300)
		Example(60)
	})
	Member("minimum_running_size", Integer, "Minimum ", func() {
		Default(1)
		Example(2)
		Description("Go health_check_error if running instance size is less than this")
	})
	Member("minimum_running_percentage", Integer, "Percentage", func() {
		Default(50)
		Example(20)
		Description("Go health_check_error if running instance size rate is less than this")
	})
})

var instanceGroupBodyRequired = []string{
	"boot_disk", "machine_type",
}

var InstanceGroupBody = Type("InstanceGroupBody", func() {
	Member("boot_disk", InstanceGroupVmDisk, "Boot disk")
	Member("machine_type", String, "GCE Machine Type", func() {
		Example("f1-micro")
	})
	Member("gpu_accelerators", InstanceGroupAccelerators, "GPU Accelerators")
	Member("preemptible", Boolean, "Use preemptible VMs")
	Member("health_check", InstanceGroupHealthCheckConfig, "Health Check setting")

	Member("instance_size", Integer, "Instance size", func() {
		Example(3)
	})
	Member("instance_size_requested", Integer, "Instance size requested", func() {
		Example(3)
	})
	Member("startup_script", String, "Startup script")

	Member("deployment_name", String, "Deployment name")

	Member("token_consumption", Integer, "Token Consumption", func() {
		Example(2)
	})
	Required(instanceGroupBodyRequired...)
})

var InstanceGroupPayload = Type("InstanceGroupPayload", func() {
	Member("pipeline_base_name", String, "Owner pipeline_base name", func() {
		// Optional
		Example("pipeline1-123")
	})
	Member("name", String, "Name", func() {
		Example("pipeline1-123-ig-456")
	})
	Member("project_id", String, "GCP Project ID", func() {
		Example("dummy-proj-999")
	})
	Member("zone", String, "GCP zone", func() {
		Example("us-central1-f")
	})
	Required("name", "project_id", "zone")

	Reference(InstanceGroupBody)
	members := []string{
		"boot_disk",
		"machine_type",
		"gpu_accelerators",
		"health_check",
		"preemptible",
		"instance_size_requested",
		"startup_script",
		"deployment_name",
		"token_consumption",
	}
	for _, m := range members {
		Member(m)
	}
	Required(instanceGroupBodyRequired...)
})

var InstanceGroup = MediaType("application/vnd.instance-group+json", func() {
	Description("instance-group")
	Reference(InstanceGroupPayload)
	attrNames := []string{
		"pipeline_base_name",
		"name",
		"project_id",
		"zone",
		"boot_disk",
		"machine_type",
		"gpu_accelerators",
		"health_check",
		"preemptible",
		"instance_size",
		"instance_size_requested",
		"startup_script",
		"deployment_name",
		"token_consumption",
	}
	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("status", String, "Instance Group Status", func() {
			Enum("construction_starting", "construction_running", "construction_error",
				"constructed",
				"health_check_error",
				"resize_starting", "resize_running", "resize_waiting",
				"destruction_starting", "destruction_running", "destruction_error", "destructed")
			Example("construction_starting")
		})
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		// Re-define instance_size because InstanceGroupPayload doesn't include it
		Attribute("instance_size", Integer, "Instance size", func() {
			Example(3)
		})
		Attribute("health_check_task_id", String, "Health check task ID")
		UseTrait(TimestampsAttrTrait)

		requiredAttrs := append([]string{"id", "status"}, attrNames...)
		Required(requiredAttrs...)
	})
	View("default", func() {
		Attribute("id")
		Attribute("status")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		UseTrait(TimestampsViewTrait)
	})
})

var _ = Resource("InstanceGroup", func() {
	BasePath("/orgs/:org_id/instance_groups")
	DefaultMedia(InstanceGroup)
	UseTrait(DefineResourceTrait)

	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
		})
		Response(OK, CollectionOf(InstanceGroup))
		UseTrait(DefaultResponseTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
		})
		Payload(InstanceGroupPayload)
		Response(Created, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:name"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("resize", func() {
		Description("Resize")
		Routing(PUT("/:name/resize"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("new_size", Integer, "New Instance Size")
			Required("new_size")
		})
		Response(OK, InstanceGroup)
		Response(Created, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("destruct", func() {
		Description("Destruct")
		Routing(PUT("/:name/destruct"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
		})
		Response(OK, InstanceGroup)
		Response(Created, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:name"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("start_health_check", func() {
		Description("Start health check")
		Routing(POST("/:name/start_health_check"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
		})
		Response(OK, InstanceGroup)
		Response(Created, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
})

var _ = Resource("InstanceGroupConstructionTask", func() {
	BasePath("/orgs/:org_id/instance_groups/:name/construction_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})

var _ = Resource("InstanceGroupDestructionTask", func() {
	BasePath("/orgs/:org_id/instance_groups/:name/destruction_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})

var _ = Resource("InstanceGroupResizingTask", func() {
	BasePath("/orgs/:org_id/instance_groups/:name/resizing_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})

var InstanceGroupHealthCheck = MediaType("application/vnd.instance-group-health-check+json", func() {
	Description("instance-group-health-check")

	Attributes(func() {
		UseTrait(IdTrait)
		Attribute("last_result", String, "Last result")
		UseTrait(TimestampsAttrTrait)
		Required("id", "created_at", "updated_at")
	})
	View("default", func() {
		Attribute("id")
		Attribute("last_result")
		UseTrait(TimestampsViewTrait)
	})
})

var _ = Resource("InstanceGroupHealthCheck", func() {
	BasePath("/orgs/:org_id/instance_groups/:name/health_checks")
	DefaultMedia(InstanceGroupHealthCheck)
	UseTrait(DefineResourceTrait)

	Action("execute", func() {
		Description("Execute health check")
		Routing(PUT("/:id"))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Param("name")
			Param("id")
		})
		Response(OK, InstanceGroupHealthCheck)             // 200 終了
		Response(Created, InstanceGroupHealthCheck)        // 201 エラーなし
		Response(Accepted, InstanceGroupHealthCheck)       // 202 致命的なエラーあり
		Response(NoContent, InstanceGroupHealthCheck)      // 204 タスクエラー
		Response(PartialContent, InstanceGroupHealthCheck) // 206 致命的でないエラーあり
		UseTrait(DefaultResponseTrait)
	})
})
