package model

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type Builder struct {
	deployer DeploymentServicer
}

func NewBuilder(ctx context.Context) (*Builder, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &Builder{deployer: deployer}, nil
}

func (b *Builder) Process(ctx context.Context, pl *InstanceGroup) (*CloudAsyncOperation, error) {
	deployment, err := b.BuildDeployment(pl)
	if err != nil {
		log.Errorf(ctx, "Failed to BuildDeployment: %v\nInstanceGroup: %v\n", err, pl)
		return nil, err
	}
	ope, err := b.deployer.Insert(ctx, pl.ProjectID, deployment)
	if err != nil {
		log.Errorf(ctx, "Failed to insert deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, deployment)
		return nil, err
	}

	log.Infof(ctx, "Built pipeline successfully %v\n", pl)

	operation := &CloudAsyncOperation{
		OwnerType:     "InstanceGroup",
		OwnerID:       pl.Id,
		ProjectId:     pl.ProjectID,
		Zone:          pl.Zone,
		Service:       "deploymentmanager",
		Name:          ope.Name,
		OperationType: ope.OperationType,
		Status:        ope.Status,
	}

	return operation, nil
}

func (b *Builder) BuildDeployment(pl *InstanceGroup) (*deploymentmanager.Deployment, error) {
	r := b.GenerateDeploymentResources(pl)
	d, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L321-L346
	c := deploymentmanager.ConfigFile{Content: string(d)}
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L1679-L1709
	tc := deploymentmanager.TargetConfiguration{Config: &c}
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L348-L434
	dm := deploymentmanager.Deployment{
		Name:   pl.Name,
		Target: &tc,
	}
	return &dm, nil
}

type (
	Resource struct {
		Type       string                 `json:"type"`
		Name       string                 `json:"name"`
		Properties map[string]interface{} `json:"properties"`
	}

	Resources struct {
		Resources []Resource `json:"resources"`
	}
)

type (
	Pubsub struct {
		Name        string
		AckDeadline int
	}
)

func (b *Builder) GenerateDeploymentResources(pl *InstanceGroup) *Resources {
	t := []Resource{}
	pubsubs := []Pubsub{
		Pubsub{Name: "job", AckDeadline: 600},
		Pubsub{Name: "progress", AckDeadline: 30},
	}
	for _, pubsub := range pubsubs {
		topic := pl.Name + "-" + pubsub.Name + "-topic"
		subscription := pl.Name + "-" + pubsub.Name + "-subscription"
		t = append(t,
			Resource{
				Type:       "pubsub.v1.topic",
				Name:       topic,
				Properties: map[string]interface{}{"topic": topic},
			},
			Resource{
				Type: "pubsub.v1.subscription",
				Name: subscription,
				Properties: map[string]interface{}{
					"subscription":       subscription,
					"topic":              fmt.Sprintf("$(ref.%s.name)", topic),
					"ackDeadlineSeconds": pubsub.AckDeadline,
				},
			},
		)
	}

	t = append(t,
		b.buildItResource(pl),
		b.buildIgmResource(pl),
	)
	return &Resources{Resources: t}
}

func (b *Builder) buildItResource(pl *InstanceGroup) Resource {
	return Resource{
		Type: "compute.v1.instanceTemplate",
		Name: pl.Name + "-it",
		Properties: map[string]interface{}{
			"zone":       pl.Zone,
			"properties": b.buildItProperties(pl),
		},
	}
}

func (b *Builder) buildItProperties(pl *InstanceGroup) map[string]interface{} {
	scheduling := map[string]interface{}{
		"preemptible": pl.Preemptible,
	}

	it_properties := map[string]interface{}{
		"machineType": pl.MachineType,
		"metadata": map[string]interface{}{
			"items": []interface{}{
				b.buildStartupScriptMetadataItem(pl),
			},
		},
		"networkInterfaces": []interface{}{
			b.buildDefaultNetwork(pl),
		},
		"scheduling": scheduling,
		"serviceAccounts": []interface{}{
			b.buildScopes(),
		},
		"disks": []interface{}{
			b.buildBootDisk(&pl.BootDisk),
		},
	}

	if pl.GpuAccelerators.Count > 0 {
		scheduling["onHostMaintenance"] = "TERMINATE"
		it_properties["guestAccelerators"] = []interface{}{
			b.buildGuestAccelerators(pl),
		}
	}

	return it_properties
}

func (b *Builder) buildGuestAccelerators(pl *InstanceGroup) map[string]interface{} {
	ga := pl.GpuAccelerators
	return map[string]interface{}{
		"acceleratorCount": float64(ga.Count),
		"acceleratorType":  ga.Type,
	}
}

func (b *Builder) buildIgmResource(pl *InstanceGroup) Resource {
	return Resource{
		Type: "compute.v1.instanceGroupManagers",
		Name: pl.Name + "-igm",
		Properties: map[string]interface{}{
			"baseInstanceName": pl.Name + "-instance",
			"instanceTemplate": "$(ref." + pl.Name + "-it.selfLink)",
			"targetSize":       pl.InstanceSizeRequested,
			"zone":             pl.Zone,
		},
	}
}

func (b *Builder) buildStartupScriptMetadataItem(pl *InstanceGroup) map[string]interface{} {
	return map[string]interface{}{
		"key":   "startup-script",
		"value": pl.StartupScript,
	}
}

func (b *Builder) buildDefaultNetwork(pl *InstanceGroup) map[string]interface{} {
	return map[string]interface{}{
		"network": "https://www.googleapis.com/compute/v1/projects/" + pl.ProjectID + "/global/networks/default",
		"accessConfigs": []interface{}{
			map[string]interface{}{
				"name": "External-IP",
				"type": "ONE_TO_ONE_NAT",
			},
		},
	}
}

func (b *Builder) buildScopes() map[string]interface{} {
	return map[string]interface{}{
		"scopes": []interface{}{
			"https://www.googleapis.com/auth/devstorage.full_control",
			"https://www.googleapis.com/auth/pubsub",
			"https://www.googleapis.com/auth/logging.write",
			"https://www.googleapis.com/auth/monitoring.write",
			"https://www.googleapis.com/auth/cloud-platform",
		},
	}
}

func (b *Builder) buildBootDisk(disk *InstanceGroupVMDisk) map[string]interface{} {
	initParams := map[string]interface{}{
		"sourceImage": disk.SourceImage,
	}
	if disk.DiskSizeGb > 0 {
		initParams["diskSizeGb"] = disk.DiskSizeGb
	}
	if disk.DiskType != "" {
		initParams["diskType"] = disk.DiskType
	}
	return map[string]interface{}{
		"deviceName":       "boot",
		"type":             "PERSISTENT",
		"boot":             true,
		"autoDelete":       true,
		"initializeParams": initParams,
	}
}