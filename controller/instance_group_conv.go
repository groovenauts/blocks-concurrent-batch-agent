package controller

import (
	"time"

	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func InstanceGroupVMDiskPayloadToModel(src *app.InstanceGroupVMDisk) model.InstanceGroupVMDisk {
	if src == nil {
		return model.InstanceGroupVMDisk{}
	}
	return model.InstanceGroupVMDisk{
		DiskSizeGb:  IntPointerToInt(src.DiskSizeGb),
		DiskType:    StringPointerToString(src.DiskType),
		SourceImage: src.SourceImage,
	}
}

func InstanceGroupVMDiskModelToMediaType(src *model.InstanceGroupVMDisk) *app.InstanceGroupVMDisk {
	if src == nil {
		return nil
	}
	return &app.InstanceGroupVMDisk{
		DiskSizeGb:  src.DiskSizeGb,
		DiskType:    src.DiskType,
		SourceImage: src.SourceImage,
	}
}

func InstanceGroupAcceleratorsPayloadToModel(src *app.InstanceGroupAccelerators) model.InstanceGroupAccelerators {
	if src == nil {
		return model.InstanceGroupAccelerators{}
	}
	return model.InstanceGroupAccelerators{
		Count: src.Count,
		Type:  src.Type,
	}
}

func InstanceGroupAcceleratorsModelToMediaType(src *model.InstanceGroupAccelerators) *app.InstanceGroupAccelerators {
	if src == nil {
		return nil
	}
	return &app.InstanceGroupAccelerators{
		Count: src.Count,
		Type:  src.Type,
	}
}

func InstanceGroupPayloadToModel(src *app.InstanceGroupPayload) model.InstanceGroup {
	if src == nil {
		return model.InstanceGroup{}
	}
	return model.InstanceGroup{
		Name:             src.Name,
		ProjectID:        src.ProjectID,
		Zone:             src.Zone,
		BootDisk:         InstanceGroupVMDiskPayloadToModel(src.BootDisk),
		MachineType:      src.MachineType,
		GpuAccelerators:  InstanceGroupAcceleratorsPayloadToModel(src.GpuAccelerators),
		Preemptible:      BoolPointerToBool(src.Preemptible),
		InstanceSize:     IntPointerToInt(src.InstanceSize),
		StartupScript:    StringPointerToString(src.StartupScript),
		DeploymentName:   StringPointerToString(src.DeploymentName),
		TokenConsumption: IntPointerToInt(src.TokenConsumption),
		// Status no payload field
		// CreatedAt no payload field
		// UpdatedAt no payload field
		// No model field for payload field "pipeline_base_id"
	}
}

func InstanceGroupModelToMediaType(src *model.InstanceGroup) *app.InstanceGroup {
	if src == nil {
		return nil
	}
	return &app.InstanceGroup{
		Name:             src.Name,
		ProjectID:        src.ProjectID,
		Zone:             src.Zone,
		BootDisk:         InstanceGroupVMDiskModelToMediaType(&src.BootDisk),
		MachineType:      src.MachineType,
		GpuAccelerators:  InstanceGroupAcceleratorsModelToMediaType(&src.GpuAccelerators),
		Preemptible:      src.Preemptible,
		InstanceSize:     src.InstanceSize,
		StartupScript:    src.StartupScript,
		Status:           src.Status,
		DeploymentName:   src.DeploymentName,
		TokenConsumption: src.TokenConsumption,
		CreatedAt:        src.CreatedAt,
		UpdatedAt:        src.UpdatedAt,
		// No field for media type field "id"
	}
}
