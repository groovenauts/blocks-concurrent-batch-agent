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

var instanceGroupBodyRequired = []string{
	"project_id", "zone", "boot_disk", "machine_type",
}

var InstanceGroupBody = Type("InstanceGroupBody", func() {
	Member("project_id", String, "GCP Project ID", func() {
		Example("dummy-proj-999")
	})
	Member("zone", String, "GCP zone", func() {
		Example("us-central1-f")
	})
	Member("boot_disk", InstanceGroupVmDisk, "Boot disk")
	Member("machine_type", String, "GCE Machine Type", func() {
		Example("f1-micro")
	})
	Member("gpu_accelerators", InstanceGroupAccelerators, "GPU Accelerators")
	Member("preemptible", Boolean, "Use preemptible VMs")

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
	Member("pipeline_base_id", String, "Owner pipeline_base id (UUID)", func() {
		// Optional
		Example("bd2d5ee3-d8be-4024-85a7-334dee9c1c88")
	})
	Member("name", String, "Name", func() {
		Example("pipeline1-123-ig-456")
	})
	Required("name")

	Reference(InstanceGroupBody)
	members := []string{
		"project_id",
		"zone",
		"boot_disk",
		"machine_type",
		"gpu_accelerators",
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
		"name",
		"project_id",
		"zone",
		"boot_disk",
		"machine_type",
		"gpu_accelerators",
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
			Enum("construction_starting", "construction_running", "construction_error", "constructed",
				"resize_starting", "resize_running",
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
	BasePath("/instance_groups")
	DefaultMedia(InstanceGroup)
	UseTrait(DefineResourceTrait)

	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Response(OK, CollectionOf(InstanceGroup))
		UseTrait(DefaultResponseTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Params(func() {
			Param("org_id", String, "Organization ID")
			Required("org_id")
		})
		Payload(InstanceGroupPayload)
		Response(Created, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("resize", func() {
		Description("Resize")
		Routing(PUT("/:id/resize"))
		Params(func() {
			Param("id")
		})
		Payload(func() {
			Member("new_size", Integer, "New Instance Size")
			Required("new_size")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("destruct", func() {
		Description("Destruct")
		Routing(PUT("/:id/destruct"))
		Params(func() {
			Param("id")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefaultResponseTrait)
	})
})

var _ = Resource("InstanceGroupConstructionTask", func() {
	BasePath("/construction_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})

var _ = Resource("InstanceGroupDestructionTask", func() {
	BasePath("/destruction_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})

var _ = Resource("InstanceGroupResizingTask", func() {
	BasePath("/resizing_tasks")
	UseTrait(DefineResourceTrait)
	UseTrait(CloudAsyncOperationResourceTrait)
})
