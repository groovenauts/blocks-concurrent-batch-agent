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

func InstanceGroupHealthCheckConfigPayloadToModel(src *app.InstanceGroupHealthCheckConfig) model.InstanceGroupHealthCheckConfig {
	if src == nil {
		return model.InstanceGroupHealthCheckConfig{}
	}
	return model.InstanceGroupHealthCheckConfig{
		Interval:                 src.Interval,
		MinimumRunningSize:       src.MinimumRunningSize,
		MinimumRunningPercentage: src.MinimumRunningPercentage,
	}
}

func InstanceGroupHealthCheckConfigModelToMediaType(src *model.InstanceGroupHealthCheckConfig) *app.InstanceGroupHealthCheckConfig {
	if src == nil {
		return nil
	}
	return &app.InstanceGroupHealthCheckConfig{
		Interval:                 src.Interval,
		MinimumRunningSize:       src.MinimumRunningSize,
		MinimumRunningPercentage: src.MinimumRunningPercentage,
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
		HealthCheck:           InstanceGroupHealthCheckConfigPayloadToModel(src.HealthCheck),
		Preemptible:           BoolPointerToBool(src.Preemptible),
		InstanceSizeRequested: IntPointerToInt(src.InstanceSizeRequested),
		StartupScript:         StringPointerToString(src.StartupScript),
		DeploymentName:        StringPointerToString(src.DeploymentName),
		TokenConsumption:      IntPointerToInt(src.TokenConsumption),
		// No model field for payload field "instance_size"
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
		HealthCheck:           InstanceGroupHealthCheckConfigModelToMediaType(&src.HealthCheck),
		Preemptible:           &src.Preemptible,
		InstanceSizeRequested: &src.InstanceSizeRequested,
		StartupScript:         &src.StartupScript,
		DeploymentName:        &src.DeploymentName,
		TokenConsumption:      &src.TokenConsumption,
		// No field for media type field "instance_size"
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
		HealthCheck:           InstanceGroupHealthCheckConfigPayloadToModel(src.HealthCheck),
		Preemptible:           BoolPointerToBool(src.Preemptible),
		InstanceSizeRequested: IntPointerToInt(src.InstanceSizeRequested),
		StartupScript:         StringPointerToString(src.StartupScript),
		DeploymentName:        StringPointerToString(src.DeploymentName),
		TokenConsumption:      IntPointerToInt(src.TokenConsumption),
		// InstanceSize no payload field
		// HealthCheckId no payload field
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
		HealthCheck:           InstanceGroupHealthCheckConfigModelToMediaType(&src.HealthCheck),
		Preemptible:           src.Preemptible,
		InstanceSizeRequested: src.InstanceSizeRequested,
		StartupScript:         src.StartupScript,
		DeploymentName:        src.DeploymentName,
		TokenConsumption:      src.TokenConsumption,
		InstanceSize:          src.InstanceSize,
		Status:                string(src.Status),
		CreatedAt:             &src.CreatedAt,
		UpdatedAt:             &src.UpdatedAt,
		// ProjectID no media type field
		// HealthCheckId no media type field
		// No field for media type field "id"
	}
}

func InstanceGroupHealthCheckModelToMediaType(src *model.InstanceGroupHealthCheck) *app.InstanceGroupHealthCheck {
	if src == nil {
		return nil
	}
	return &app.InstanceGroupHealthCheck{
		LastResult: &src.LastResult,
		CreatedAt:  src.CreatedAt,
		UpdatedAt:  src.UpdatedAt,
		// No field for media type field "id"
	}
}
