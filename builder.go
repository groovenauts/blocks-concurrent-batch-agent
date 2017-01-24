package pipeline

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"encoding/json"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/deploymentmanager/v2"
)

type Builder struct {
}

func (b *Builder) Process(ctx context.Context, pl *Pipeline) error {
	pl.Props.Status = building
	err := pl.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'building': %v\npl: %v\n", err, pl)
		return err
	}

	log.Debugf(ctx, "Building pipeline %v\n", pl.Props)
	// See the "Examples" below "Response"
	//   https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/insert#response
	hc, err := google.DefaultClient(ctx, deploymentmanager.CloudPlatformScope)
	if err != nil {
		log.Errorf(ctx, "Failed to get google.DefaultClient: %v\n", err)
		return err
	}
	c, err := deploymentmanager.New(hc)
	if err != nil {
		log.Errorf(ctx, "Failed to get deploymentmanager.New(hc): %v\nhc: %v\n", err, hc)
		return err
	}
	deployment, err := b.BuildDeployment(&pl.Props)
	if err != nil {
		log.Errorf(ctx, "Failed to BuildDeployment: %v\nProps: %v\n", err, pl.Props)
		return err
	}
	c.Deployments.Insert(pl.Props.ProjectID, deployment).Context(ctx).Do()
	log.Infof(ctx, "Built pipeline successfully %v\n", pl.Props)

	pl.Props.Status = opened
	err = pl.update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline status to 'opened': %v\npl: %v\n", err, pl)
		return err
	}

	return nil
}

func (b *Builder) BuildDeployment(plp *PipelineProps) (*deploymentmanager.Deployment, error) {
	r := b.GenerateDeploymentResources(plp)
	d, err := json.Marshal(r)
	if err != nil { return nil, err }
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L321-L346
	c := deploymentmanager.ConfigFile{Content: string(d)}
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L1679-L1709
	tc := deploymentmanager.TargetConfiguration{Config: &c}
	// https://github.com/google/google-api-go-client/blob/master/deploymentmanager/v2/deploymentmanager-gen.go#L348-L434
	dm := deploymentmanager.Deployment{
		Name: plp.Name,
		Target: &tc,
	}
	return &dm, nil
}

type (
	Resource struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Properties map[string]interface{} `json:"properties"`
	}

	Resources struct {
		Resources []Resource `json:"resources"`
	}
)

type (
	Pubsub struct {
		Name string
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
				Type: "pubsub.v1.topic",
				Name: topic,
				Properties: map[string]interface{}{ "topic": topic },
			},
			Resource{
				Type: "pubsub.v1.subscription",
				Name: subscription,
				Properties: map[string]interface{}{
					"subscription": subscription,
					"topic": fmt.Sprintf("$(ref.%s.name)", topic),
					"ackDeadlineSeconds": pubsub.AckDeadline,
				},
			},
		)
	}

	startup_script :=
		fmt.Sprintf("for i in {1..%v}; do", plp.ContainerSize) +
		" docker run -d" +
		" -e BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=$(ref.pipeline01-job-subscription.name)" +
		" -e BLOCKS_BATCH_PROGRESS_TOPIC=$(ref.pipeline01-progress-topic.name)" +
		" " + plp.ContainerName +
		" " + plp.Command +
		" ; done"

	t = append(t,
		Resource{
			Type: "compute.v1.instanceTemplate",
			Name: plp.Name + "-it",
			Properties: map[string]interface{}{
        "zone": plp.Zone,
        "properties": map[string]interface{}{
          "machineType": plp.MachineType,
					"metadata": map[string]interface{}{
						"items": []interface{} {
							map[string]interface{}{
								"key": "startup-script",
								"value": startup_script,
							},
						},
					},
          "networkInterfaces": []interface{}{
            map[string]interface{}{
              "network": "https://www.googleapis.com/compute/v1/projects/" +plp.ProjectID+ "/global/networks/default",
              "accessConfigs": []interface{}{
                map[string]interface{}{
                  "name": "External-IP",
                  "type": "ONE_TO_ONE_NAT",
                },
              },
            },
          },
					"serviceAccounts": []interface{}{
						map[string]interface {}{
							"scopes": []interface{}{
								"https://www.googleapis.com/auth/devstorage.full_control",
								"https://www.googleapis.com/auth/pubsub",
							},
						},
					},
          "disks": []interface{}{
            map[string]interface{}{
              "deviceName": "boot",
              "type": "PERSISTENT",
              "boot": true,
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
        "instanceTemplate": "$(ref.pipeline01-it.selfLink)",
        "targetSize": plp.TargetSize,
        "zone": plp.Zone,
      },
    },
	)
	return &Resources{Resources: t}
}
