package model

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type PipelineOpener struct {
	deployer DeploymentServicer
}

func NewPipelineOpener(ctx context.Context) (*PipelineOpener, error) {
	deployer, err := DefaultDeploymentServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &PipelineOpener{deployer: deployer}, nil
}

func (b *PipelineOpener) Process(ctx context.Context, pl *InstanceGroup) (*CloudAsyncOperation, error) {
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

func (b *PipelineOpener) BuildDeployment(pl *InstanceGroup) (*deploymentmanager.Deployment, error) {
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
	Pubsub struct {
		Name        string
		AckDeadline int
	}
)

func (b *PipelineOpener) GenerateDeploymentResources(pl *InstanceGroup) *Resources {
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

	return &Resources{Resources: t}
}
