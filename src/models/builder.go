package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

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

func (b *Builder) Process(ctx context.Context, pl *Pipeline) (*PipelineOperation, error) {
	err := pl.LoadOrganization(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to load Organization for Pipeline: %v\npl: %v\n", err, pl)
		return nil, err
	}

	err = pl.StartBuilding(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'building': %v\npl: %v\n", err, pl)
		return nil, err
	}

	deployment, err := b.BuildDeployment(pl)
	if err != nil {
		log.Errorf(ctx, "Failed to BuildDeployment: %v\nPipeline: %v\n", err, pl)
		return nil, err
	}
	ope, err := b.deployer.Insert(ctx, pl.ProjectID, deployment)
	if err != nil {
		log.Errorf(ctx, "Failed to insert deployment %v\nproject: %v deployment: %v\n", err, pl.ProjectID, deployment)
		return nil, err
	}

	log.Infof(ctx, "Built pipeline successfully %v\n", pl)

	err = pl.StartDeploying(ctx, deployment.Name)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline deployment name to %v: %v\npl: %v\n", ope.Name, err, pl)
		return nil, err
	}

	operation := &PipelineOperation{
		Pipeline:      pl,
		ProjectID:     pl.ProjectID,
		Zone:          pl.Zone,
		Service:       "deploymentmanager",
		Name:          ope.Name,
		OperationType: ope.OperationType,
		Status:        ope.Status,
		Logs: []OperationLog{
			OperationLog{CreatedAt: time.Now(), Message: "Start"},
		},
	}
	err = operation.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create PipelineOperation: %v because of %v\n", operation, err)
		return nil, err
	}

	return operation, nil
}

func (b *Builder) BuildDeployment(pl *Pipeline) (*deploymentmanager.Deployment, error) {
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

func (b *Builder) GenerateDeploymentResources(pl *Pipeline) *Resources {
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

func (b *Builder) buildItResource(pl *Pipeline) Resource {
	return Resource{
		Type: "compute.v1.instanceTemplate",
		Name: pl.Name + "-it",
		Properties: map[string]interface{}{
			"zone":       pl.Zone,
			"properties": b.buildItProperties(pl),
		},
	}
}

func (b *Builder) buildItProperties(pl *Pipeline) map[string]interface{} {
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

func (b *Builder) buildGuestAccelerators(pl *Pipeline) map[string]interface{} {
	ga := pl.GpuAccelerators
	return map[string]interface{}{
		"acceleratorCount": float64(ga.Count),
		"acceleratorType":  ga.Type,
	}
}

func (b *Builder) buildIgmResource(pl *Pipeline) Resource {
	return Resource{
		Type: "compute.v1.instanceGroupManagers",
		Name: pl.Name + "-igm",
		Properties: map[string]interface{}{
			"baseInstanceName": pl.Name + "-instance",
			"instanceTemplate": "$(ref." + pl.Name + "-it.selfLink)",
			"targetSize":       pl.TargetSize,
			"zone":             pl.Zone,
		},
	}
}

func (b *Builder) buildStartupScriptMetadataItem(pl *Pipeline) map[string]interface{} {
	startup_script := b.buildStartupScript(pl)
	return map[string]interface{}{
		"key":   "startup-script",
		"value": startup_script,
	}
}

func (b *Builder) buildDefaultNetwork(pl *Pipeline) map[string]interface{} {
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

func (b *Builder) buildBootDisk(disk *PipelineVmDisk) map[string]interface{} {
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

const GcrHostPatternBase = `\A[^/]*gcr.io`

var (
	CosCloudProjectRegexp   = regexp.MustCompile(`/projects/cos-cloud/`)
	GcrContainerImageRegexp = regexp.MustCompile(GcrHostPatternBase + `\/`)
	GcrImageHostRegexp      = regexp.MustCompile(GcrHostPatternBase)
)

const StackdriverAgentCommand = "docker run -d -e MONITOR_HOST=true -v /proc:/mnt/proc:ro --privileged wikiwi/stackdriver-agent"

func (b *Builder) buildStartupScript(pl *Pipeline) string {
	r := []string{StartupScriptHeader}

	docker := "docker"
	if pl.GpuAccelerators.Count > 0 {
		r = append(r,
			b.buildInstallCuda(pl),
			b.buildInstallDocker(pl),
			b.buildInstallNvidiaDocker(pl),
		)
		docker = "nvidia-docker"
	}

	usingGcr :=
		CosCloudProjectRegexp.MatchString(pl.BootDisk.SourceImage) &&
			GcrContainerImageRegexp.MatchString(pl.ContainerName)
	if usingGcr {
		docker = docker + " --config /home/chronos/.docker"
		host := GcrImageHostRegexp.FindString(pl.ContainerName)
		r = append(r,
			"METADATA=http://metadata.google.internal/computeMetadata/v1",
			"SVC_ACCT=$METADATA/instance/service-accounts/default",
			"ACCESS_TOKEN=$(curl -H 'Metadata-Flavor: Google' $SVC_ACCT/token | cut -d'\"' -f 4)",
			"INITIAL_INTERVAL=60 with_backoff "+docker+" login -e 1234@5678.com -u _token -p $ACCESS_TOKEN https://"+host,
		)
	}

	if pl.StackdriverAgent {
		r = append(r, StackdriverAgentCommand)
	}

	docker_run_parts := []string{
		docker + " run -d",
		"-e PROJECT=" + pl.ProjectID,
		"-e DOCKER_HOSTNAME=$(hostname)",
		"-e PIPELINE=" + pl.Name,
		"-e ZONE=" + pl.Zone,
		"-e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref." + pl.Name + "-job-subscription.name)",
		"-e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref." + pl.Name + "-progress-topic.name)",
		pl.ContainerName,
		pl.Command,
	}

	r = append(r,
		"INITIAL_INTERVAL=600 with_backoff "+docker+" pull "+pl.ContainerName,
		fmt.Sprintf("for i in {1..%v}; do", pl.ContainerSize),
		"  "+strings.Join(docker_run_parts, " \\\n    "),
		"done",
	)
	return strings.Join(r, "\n")
}

func (b *Builder) buildInstallCuda(pl *Pipeline) string {
	return `
if ! dpkg-query -W cuda; then
   curl -O http://developer.download.nvidia.com/compute/cuda/repos/ubuntu1604/x86_64/cuda-repo-ubuntu1604_8.0.61-1_amd64.deb
   dpkg -i ./cuda-repo-ubuntu1604_8.0.61-1_amd64.deb
   apt-get update
   apt-get -y install cuda
fi
nvidia-smi
`
}

func (b *Builder) buildInstallDocker(pl *Pipeline) string {
	return `
apt-get update
apt-get -y install \
     apt-transport-https \
     ca-certificates \
     curl \
     software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
apt-key fingerprint 0EBFCD88
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update
apt-get -y install docker-ce
docker run hello-world
`
}

func (b *Builder) buildInstallNvidiaDocker(pl *Pipeline) string {
	return `
wget -P /tmp https://github.com/NVIDIA/nvidia-docker/releases/download/v1.0.1/nvidia-docker_1.0.1-1_amd64.deb
dpkg -i /tmp/nvidia-docker*.deb && rm /tmp/nvidia-docker*.deb
nvidia-docker run --rm nvidia/cuda nvidia-smi
`
}
