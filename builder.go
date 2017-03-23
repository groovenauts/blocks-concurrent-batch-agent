package pipeline

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
	"regexp"
)

type Builder struct {
	deployer DeploymentServicer
}

func (b *Builder) Process(ctx context.Context, pl *Pipeline) error {
	pl.Props.Status = building
	err := pl.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'building': %v\npl: %v\n", err, pl)
		return err
	}

	deployment, err := b.BuildDeployment(&pl.Props)
	if err != nil {
		log.Errorf(ctx, "Failed to BuildDeployment: %v\nProps: %v\n", err, pl.Props)
		return err
	}
	ope, err := b.deployer.Insert(ctx, pl.Props.ProjectID, deployment)
	if err != nil {
		log.Errorf(ctx, "Failed to insert deployment %v\nproject: %v deployment: %v\n", err, pl.Props.ProjectID, deployment)
		return err
	}

	log.Infof(ctx, "Built pipeline successfully %v\n", pl.Props)

	pl.Props.Status = deploying
	pl.Props.DeploymentName = deployment.Name
	pl.Props.DeployingOperationName = ope.Name
	err = pl.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline deployment name to %v: %v\npl: %v\n", ope.Name, err, pl)
		return err
	}

	return nil
}

func (b *Builder) BuildDeployment(plp *PipelineProps) (*deploymentmanager.Deployment, error) {
	r := b.GenerateDeploymentResources(plp)
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
		Name:   plp.Name,
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

func (b *Builder) GenerateDeploymentResources(plp *PipelineProps) *Resources {
	t := []Resource{}
	pubsubs := []Pubsub{
		Pubsub{Name: "job", AckDeadline: 600},
		Pubsub{Name: "progress", AckDeadline: 30},
	}
	for _, pubsub := range pubsubs {
		topic := plp.Name + "-" + pubsub.Name + "-topic"
		subscription := plp.Name + "-" + pubsub.Name + "-subscription"
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

	startup_script := b.buildStartupScript(plp)

	t = append(t,
		Resource{
			Type: "compute.v1.instanceTemplate",
			Name: plp.Name + "-it",
			Properties: map[string]interface{}{
				"zone": plp.Zone,
				"properties": map[string]interface{}{
					"machineType": plp.MachineType,
					"metadata": map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{
								"key":   "startup-script",
								"value": startup_script,
							},
						},
					},
					"networkInterfaces": []interface{}{
						map[string]interface{}{
							"network": "https://www.googleapis.com/compute/v1/projects/" + plp.ProjectID + "/global/networks/default",
							"accessConfigs": []interface{}{
								map[string]interface{}{
									"name": "External-IP",
									"type": "ONE_TO_ONE_NAT",
								},
							},
						},
					},
					"serviceAccounts": []interface{}{
						map[string]interface{}{
							"scopes": []interface{}{
								"https://www.googleapis.com/auth/devstorage.full_control",
								"https://www.googleapis.com/auth/pubsub",
							},
						},
					},
					"disks": []interface{}{
						map[string]interface{}{
							"deviceName": "boot",
							"type":       "PERSISTENT",
							"boot":       true,
							"autoDelete": true,
							"initializeParams": map[string]interface{}{
								"sourceImage": plp.SourceImage,
							},
						},
					},
				},
			},
		},
		Resource{
			Type: "compute.v1.instanceGroupManagers",
			Name: plp.Name + "-igm",
			Properties: map[string]interface{}{
				"baseInstanceName": plp.Name + "-instance",
				"instanceTemplate": "$(ref." + plp.Name + "-it.selfLink)",
				"targetSize":       plp.TargetSize,
				"zone":             plp.Zone,
			},
		},
	)
	return &Resources{Resources: t}
}

const GcrHostPatternBase = `\A[^/]*gcr.io`

var (
	CosCloudProjectRegexp   = regexp.MustCompile(`/projects/cos-cloud/`)
	GcrContainerImageRegexp = regexp.MustCompile(GcrHostPatternBase + `\/`)
	GcrImageHostRegexp      = regexp.MustCompile(GcrHostPatternBase)
)

func (b *Builder) buildStartupScript(plp *PipelineProps) string {
	r :=
		fmt.Sprintf("for i in {1..%v}; do", plp.ContainerSize) +
			" docker run -d" +
			" -e PROJECT=" + plp.ProjectID +
			" -e PIPELINE=" + plp.Name +
			" -e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref." + plp.Name + "-job-subscription.name)" +
			" -e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref." + plp.Name + "-progress-topic.name)" +
			" " + plp.ContainerName +
			" " + plp.Command +
			" ; done"
	usingGcr :=
		CosCloudProjectRegexp.MatchString(plp.SourceImage) &&
			GcrContainerImageRegexp.MatchString(plp.ContainerName)
	if usingGcr {
		host := GcrImageHostRegexp.FindString(plp.ContainerName)
		r =
			"METADATA=http://metadata.google.internal/computeMetadata/v1\n" +
				"SVC_ACCT=$METADATA/instance/service-accounts/default\n" +
				"ACCESS_TOKEN=$(curl -H 'Metadata-Flavor: Google' $SVC_ACCT/token | cut -d'\"' -f 4)\n" +
				"docker login -e 1234@5678.com -u _token -p $ACCESS_TOKEN https://" + host + "\n" +
				"docker pull " + plp.ContainerName + "\n" +
				r
	}
	return r
}
