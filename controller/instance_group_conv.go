package controller

import (
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
		DiskSizeGb:  &src.DiskSizeGb,
		DiskType:    &src.DiskType,
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

func InstanceGroupBodyPayloadToModel(src *app.InstanceGroupBody) model.InstanceGroupBody {
	if src == nil {
		return model.InstanceGroupBody{}
	}
	return model.InstanceGroupBody{
		BootDisk:              InstanceGroupVMDiskPayloadToModel(src.BootDisk),
		MachineType:           src.MachineType,
		GpuAccelerators:       InstanceGroupAcceleratorsPayloadToModel(src.GpuAccelerators),
		Preemptible:           BoolPointerToBool(src.Preemptible),
		InstanceSizeRequested: IntPointerToInt(src.InstanceSizeRequested),
		InstanceSize:          IntPointerToInt(src.InstanceSize),
		StartupScript:         StringPointerToString(src.StartupScript),
		DeploymentName:        StringPointerToString(src.DeploymentName),
		TokenConsumption:      IntPointerToInt(src.TokenConsumption),
		// ProjectID no payload field
		// Zone no payload field
		// Status no payload field
	}
}

func InstanceGroupBodyModelToMediaType(src *model.InstanceGroupBody) *app.InstanceGroupBody {
	if src == nil {
		return nil
	}
	return &app.InstanceGroupBody{
		BootDisk:              InstanceGroupVMDiskModelToMediaType(&src.BootDisk),
		MachineType:           src.MachineType,
		GpuAccelerators:       InstanceGroupAcceleratorsModelToMediaType(&src.GpuAccelerators),
		Preemptible:           &src.Preemptible,
		InstanceSizeRequested: &src.InstanceSizeRequested,
		InstanceSize:          &src.InstanceSize,
		StartupScript:         &src.StartupScript,
		DeploymentName:        &src.DeploymentName,
		TokenConsumption:      &src.TokenConsumption,
		// ProjectID no media type field
		// Zone no media type field
		// Status no media type field
	}
}

func InstanceGroupPayloadToModel(src *app.InstanceGroupPayload) model.InstanceGroup {
	if src == nil {
		return model.InstanceGroup{}
	}
	return model.InstanceGroup{
		Name:                  src.Name,
		ProjectID:             src.ProjectID,
		Zone:                  src.Zone,
		BootDisk:              InstanceGroupVMDiskPayloadToModel(src.BootDisk),
		MachineType:           src.MachineType,
		GpuAccelerators:       InstanceGroupAcceleratorsPayloadToModel(src.GpuAccelerators),
		Preemptible:           BoolPointerToBool(src.Preemptible),
		InstanceSizeRequested: IntPointerToInt(src.InstanceSizeRequested),
		StartupScript:         StringPointerToString(src.StartupScript),
		DeploymentName:        StringPointerToString(src.DeploymentName),
		TokenConsumption:      IntPointerToInt(src.TokenConsumption),
		// InstanceSize no payload field
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
		Name:                  src.Name,
		Zone:                  src.Zone,
		BootDisk:              InstanceGroupVMDiskModelToMediaType(&src.BootDisk),
		MachineType:           src.MachineType,
		GpuAccelerators:       InstanceGroupAcceleratorsModelToMediaType(&src.GpuAccelerators),
		Preemptible:           src.Preemptible,
		InstanceSizeRequested: src.InstanceSizeRequested,
		InstanceSize:          src.InstanceSize,
		StartupScript:         src.StartupScript,
		Status:                string(src.Status),
		DeploymentName:        src.DeploymentName,
		TokenConsumption:      src.TokenConsumption,
		CreatedAt:             &src.CreatedAt,
		UpdatedAt:             &src.UpdatedAt,
		// ProjectID no media type field
		// No field for media type field "id"
	}
}
