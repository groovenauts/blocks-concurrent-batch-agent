package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func PipelineVmDiskPayloadToModel(src *app.PipelineVMDisk) model.PipelineVmDisk {
	if src == nil {
		return model.PipelineVmDisk{}
	}
	return model.PipelineVmDisk{
		DiskSizeGb:  src.DiskSizeGb,
		DiskType:    src.DiskType,
		SourceImage: src.SourceImage,
	}
}

func PipelineVmDiskModelToMediaType(src *model.PipelineVmDisk) *app.PipelineVMDisk {
	if src == nil {
		return &app.PipelineVMDisk{}
	}
	return &app.PipelineVMDisk{
		DiskSizeGb:  src.DiskSizeGb,
		DiskType:    src.DiskType,
		SourceImage: src.SourceImage,
	}
}

func AcceleratorsPayloadToModel(src *app.Accelerators) model.Accelerators {
	if src == nil {
		return model.Accelerators{}
	}
	return model.Accelerators{
		Count: src.Count,
		Type:  src.Type,
	}
}

func AcceleratorsModelToMediaType(src *model.Accelerators) *app.Accelerators {
	if src == nil {
		return &app.Accelerators{}
	}
	return &app.Accelerators{
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
		BootDisk:         PipelineVmDiskPayloadToModel(src.BootDisk),
		MachineType:      src.MachineType,
		GpuAccelerators:  AcceleratorsPayloadToModel(src.GpuAccelerators),
		Preemptible:      BoolPointerToBool(src.Preemptible),
		InstanceSize:     IntPointerToInt(src.InstanceSize),
		StartupScript:    StringPointerToString(src.StartupScript),
		DeploymentName:   StringPointerToString(src.DeploymentName),
		TokenConsumption: IntPointerToInt(src.TokenConsumption),
	}
}

func InstanceGroupModelToMediaType(src *model.InstanceGroup) *app.InstanceGroup {
	if src == nil {
		return nil
	}
	return &app.InstanceGroup{
		ID:               src.Id,
		Name:             src.Name,
		ProjectID:        src.ProjectID,
		Zone:             src.Zone,
		BootDisk:         PipelineVmDiskModelToMediaType(&src.BootDisk),
		MachineType:      src.MachineType,
		GpuAccelerators:  AcceleratorsModelToMediaType(&src.GpuAccelerators),
		Preemptible:      src.Preemptible,
		InstanceSize:     src.InstanceSize,
		StartupScript:    src.StartupScript,
		Status:           src.Status,
		DeploymentName:   src.DeploymentName,
		TokenConsumption: src.TokenConsumption,
	}
}
