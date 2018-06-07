package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func PipelineVmDiskAppToModel(app *app.PipelineVMDisk) model.PipelineVmDisk {
	if app == nil {
		return model.PipelineVmDisk{}
	}
	return model.PipelineVmDisk{
		DiskSizeGb:  app.DiskSizeGb,
		DiskType:    app.DiskType,
		SourceImage: app.SourceImage,
	}
}

func AcceleratorsAppToModel(app *app.Accelerators) model.Accelerators {
	if app == nil {
		return model.Accelerators{}
	}
	return model.Accelerators{
		Count: app.Count,
		Type:  app.Type,
	}
}

func InstanceGroupAppToModel(app *app.InstanceGroup) model.InstanceGroup {
	return model.InstanceGroup{
		Name:             app.Name,
		ProjectID:        app.ProjectID,
		Zone:             app.Zone,
		BootDisk:         PipelineVmDiskAppToModel(app.BootDisk),
		MachineType:      app.MachineType,
		GpuAccelerators:  AcceleratorsAppToModel(app.GpuAccelerators),
		Preemptible:      app.Preemptible,
		StackdriverAgent: app.StackdriverAgent,
		InstanceSize:     app.InstanceSize,
		StartupScript:    app.StartupScript,
		Status:           app.Status,
		DeploymentName:   app.DeploymentName,
		TokenConsumption: app.TokenConsumption,
	}
}
