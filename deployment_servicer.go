package pipeline

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/log"
)

type (
	// The interface to make mock of DeploymentsService.
	// See https://godoc.org/google.golang.org/api/deploymentmanager/v2#DeploymentsService
	DeploymentServicer interface {
		Delete(project string, deployment string) *deploymentmanager.DeploymentsDeleteCall
		Get(project string, deployment string) *deploymentmanager.DeploymentsGetCall
		Insert(project string, deployment *deploymentmanager.Deployment) *deploymentmanager.DeploymentsInsertCall
	}
)

func DefaultDeploymentServicer(ctx context.Context) (*deploymentmanager.DeploymentsService, error) {
	// https://cloud.google.com/deployment-manager/docs/reference/latest/deployments/insert#examples
	hc, err := google.DefaultClient(ctx, deploymentmanager.CloudPlatformScope)
	if err != nil {
		log.Errorf(ctx, "Failed to get google.DefaultClient: %v\n", err)
		return nil, err
	}
	c, err := deploymentmanager.New(hc)
	if err != nil {
		log.Errorf(ctx, "Failed to get deploymentmanager.New(hc): %v\nhc: %v\n", err, hc)
		return nil, err
	}
	return c.Deployments, nil
}
