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
		Delete(ctx context.Context, project string, deployment string) (*deploymentmanager.Operation, error)
		Get(ctx context.Context, project string, deployment string) (*deploymentmanager.Deployment, error)
		Insert(ctx context.Context, project string, deployment *deploymentmanager.Deployment) (*deploymentmanager.Operation, error)
		GetOperation(ctx context.Context, project string, operation string) (*deploymentmanager.Operation, error)
	}
)

func DefaultDeploymentServicer(ctx context.Context) (DeploymentServicer, error) {
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
	return &DeploymentServiceWrapper{service: c.Deployments, opeService: c.Operations}, nil
}

type DeploymentServiceWrapper struct {
	service *deploymentmanager.DeploymentsService
	opeService *deploymentmanager.OperationsService
}

func (w *DeploymentServiceWrapper) Delete(ctx context.Context, project string, deployment string) (*deploymentmanager.Operation, error) {
	return w.service.Delete(project, deployment).Context(ctx).Do()
}

func (w *DeploymentServiceWrapper) Get(ctx context.Context, project string, deployment string) (*deploymentmanager.Deployment, error) {
	return w.service.Get(project, deployment).Context(ctx).Do()
}

func (w *DeploymentServiceWrapper) Insert(ctx context.Context, project string, deployment *deploymentmanager.Deployment) (*deploymentmanager.Operation, error) {
	return w.service.Insert(project, deployment).Context(ctx).Do()
}

func (w *DeploymentServiceWrapper) GetOperation(ctx context.Context, project string, operation string) (*deploymentmanager.Operation, error) {
	return w.opeService.Get(project, operation).Context(ctx).Do()
}
