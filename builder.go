package pipeline

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"gopkg.in/yaml.v2"
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
	r := b.GenerateDeploymentResources(plp.ProjectID, plp.Name)
	d, err := yaml.Marshal(r)
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
		Type string `yaml:"type"`
		Name string `yaml:"name"`
		Properties map[string]interface{} `yaml:"properties"`
	}

	Resources struct {
		Resources []Resource `yaml:"resources"`
	}
)

type (
	Pubsub struct {
		Name string
		AckDeadline int
	}
)

func (b *Builder) GenerateDeploymentResources(project, name string) *Resources {
	t := []Resource{}
	pubsubs := []Pubsub{
		Pubsub{Name: "job", AckDeadline: 600},
		Pubsub{Name: "progress", AckDeadline: 30},
	}
	for _, pubsub := range pubsubs {
		topic := name + "-" + pubsub.Name + "-topic"
		subscription := name + "-" + pubsub.Name + "-subscription"
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
	return &Resources{Resources: t}
}
