package design

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
)

var PipelineVmDisk = Type("PipelineVmDisk", func() {
	// Member("disk_name", String, "Disk name")
	Member("disk_size_gb", Integer, "Disk size", func() {
		Default(0)
		Example(50)
	})
	Member("disk_type", String, "Disk type", func() {
		Default("")
		Example("projects/dummy-proj-999/zones/asia-east1-a/diskTypes/pd-standard")
	})
	Member("source_image", String, "Source image", func() {
		Default("")
		Example("https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable")
	})
	Required("source_image")
})

var Accelerators = Type("Accelerators", func() {
	Member("count", Integer, "Count", func() {
		Default(0)
		Example(2)
	})
	Member("type", String, "Type", func() {
		Default("")
		Example("nvidia-tesla-p100")
	})
	Required("count", "type")
})

var InstanceGroupPayload = Type("InstanceGroupPayload", func() {
	Member("name", String, "Name", func() {
		Example("instancegroup1")
	})
	Member("project_id", String, "GCP Project ID", func() {
		Example("dummy-proj-999")
	})
	Member("zone", String, "GCP zone", func() {
		Example("us-central1-f")
	})
	Member("boot_disk", PipelineVmDisk, "Boot disk")
	Member("machine_type", String, "GCE Machine Type", func() {
		Example("f1-micro")
	})
	Member("gpu_accelerators", Accelerators, "GPU Accelerators")
	Member("preemptible", Boolean, "Use preemptible VMs")

	Member("instance_size", Integer, "Instance size", func() {
		Example(3)
	})
	Member("startup_script", String, "Startup script")

	Member("status", String, "Status", func() {
		Enum("constructing", "constructing_error", "constructed", "resizing", "destructing", "destructing_error", "destructed")
	})
	Member("deployment_name", String, "Deployment name")

	Member("token_consumption", Integer, "Token Consumption", func() {
		Example(2)
	})

	Required("name", "project_id", "zone", "boot_disk", "machine_type")
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
		"startup_script",
		"status",
		"deployment_name",
		"token_consumption",
	}
	outputAttrs := []string{
		"created_at",
		"updated_at",
	}
	Attributes(func() {
		Attribute("id", String, "ID", func() {
			Example("bhJifmNvbmN1cnJlbnQtYmF0Y2hyMAsSDU9yZ2FuaXphdGlvbmMYgICAgJK2lgoMCxIJUGlwZWxpbmVzGICAgIDAnIIKDX")
		})
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		Attribute("created_at", DateTime, "Datetime created")
		Attribute("updated_at", DateTime, "Datetime updated")

		requiredAttrs := append(append([]string{"id"}, attrNames...), outputAttrs...)
		Required(requiredAttrs...)
	})
	View("default", func() {
		Attribute("id")
		for _, attrName := range attrNames {
			Attribute(attrName)
		}
		for _, attrName := range outputAttrs {
			Attribute(attrName)
		}
	})
})

var _ = Resource("IntanceGroup", func() {
	BasePath("/instance_groups")
	DefaultMedia(InstanceGroup)
	Action("list", func() {
		Description("list")
		Routing(GET(""))
		Response(OK, CollectionOf(InstanceGroup))
		UseTrait(DefineTrait)
	})
	Action("create", func() {
		Description("create")
		Routing(POST(""))
		Payload(InstanceGroupPayload)
		Response(Created, InstanceGroup)
		UseTrait(DefineTrait)
	})
	Action("show", func() {
		Description("show")
		Routing(GET("/:id"))
		Params(func() {
			Param("id")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefineTrait)
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
		UseTrait(DefineTrait)
	})
	Action("destruct", func() {
		Description("Destruct")
		Routing(PUT("/:id/restruct"))
		Params(func() {
			Param("id")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefineTrait)
	})
	Action("delete", func() {
		Description("delete")
		Routing(DELETE("/:id"))
		Params(func() {
			Param("id")
			Required("id")
		})
		Response(OK, InstanceGroup)
		UseTrait(DefineTrait)
	})
})
