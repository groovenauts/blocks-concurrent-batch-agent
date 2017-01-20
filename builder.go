package pipeline

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	// "gopkg.in/yaml.v2"
)

type Builder struct {
}

func (b *Builder) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Building pipeline %v\n", pl)
	return nil
}

type (
	Metadata struct {
		DependsOn []string `yaml:"dependsOn"`
	}

	Resource struct {
		Type string `yaml:"type"`
		Name string `yaml:"name"`
		Properties map[string]string `yaml:"properties"`
		Metadata Metadata `yaml:"metadata"`
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
				Properties: map[string]string{ "topic": topic },
			},
			Resource{
				Type: "pubsub.v1.subscription",
				Name: subscription,
				Properties: map[string]string{
					"subscription": subscription,
					"topic": fmt.Sprintf("projects/%s/topics/%s", project, topic),
					"ackDeadlineSeconds": fmt.Sprintf("%v", pubsub.AckDeadline),
				},
				Metadata: Metadata{
					DependsOn: []string{ topic },
				},
			},
		)
	}
	return &Resources{Resources: t}
}
