package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine/aetest"
	//"google.golang.org/appengine/log"
)

type TestDeployerRunning struct{}

func (d *TestDeployerRunning) Delete(ctx context.Context, project string, deployment string) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerRunning) Insert(ctx context.Context, project string, deployment *deploymentmanager.Deployment) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerRunning) Get(ctx context.Context, project string, deployment string) (*deploymentmanager.Deployment, error) {
	return &deploymentmanager.Deployment{
		Operation: &deploymentmanager.Operation{
			Status: "RUNNING",
		},
	}, nil
}

type TestDeployerOK struct{}

func (d *TestDeployerOK) Delete(ctx context.Context, project string, deployment string) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerOK) Insert(ctx context.Context, project string, deployment *deploymentmanager.Deployment) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerOK) Get(ctx context.Context, project string, deployment string) (*deploymentmanager.Deployment, error) {
	return &deploymentmanager.Deployment{
		Operation: &deploymentmanager.Operation{
			Status: "DONE",
			Error:  nil,
		},
	}, nil
}

type TestDeployerError struct{}

func (d *TestDeployerError) Delete(ctx context.Context, project string, deployment string) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerError) Insert(ctx context.Context, project string, deployment *deploymentmanager.Deployment) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerError) Get(ctx context.Context, project string, deployment string) (*deploymentmanager.Deployment, error) {
	return &deploymentmanager.Deployment{
		Operation: &deploymentmanager.Operation{
			Status: "DONE",
			Error: &deploymentmanager.OperationError{
				Errors: []*deploymentmanager.OperationErrorErrors{
					&deploymentmanager.OperationErrorErrors{
						Code:     "999",
						Location: "Somewhere",
						Message:  "Something wrong",
					},
				},
			},
		},
	}, nil
}

func TestRefresherProcess(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.NoError(t, err)
	defer done()

	type Expection struct {
		status   Status
		deployer DeploymentServicer
		errors   []DeploymentError
	}

	expections := []Expection{
		Expection{
			status:   deploying,
			deployer: &TestDeployerRunning{},
			errors:   nil,
		},
		Expection{
			status:   opened,
			deployer: &TestDeployerOK{},
			errors:   nil,
		},
		Expection{
			status:   broken,
			deployer: &TestDeployerError{},
			errors: []DeploymentError{
				DeploymentError{
					Code:     "999",
					Location: "Somewhere",
					Message:  "Something wrong",
				},
			},
		},
	}

	for _, expection := range expections {
		plp := PipelineProps{
			Name:           "pipeline01",
			ProjectID:      proj,
			Zone:           "us-central1-f",
			SourceImage:    "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			MachineType:    "f1-micro",
			TargetSize:     2,
			ContainerSize:  2,
			ContainerName:  "groovenauts/batch_type_iot_example:0.3.1",
			Command:        "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
			DeploymentName: "pipeline01",
			Status:         deploying,
		}
		pl, err := CreatePipeline(ctx, &plp)

		r := &Refresher{deployer: expection.deployer}
		err = r.Process(ctx, pl)
		assert.NoError(t, err)
		pl2, err := FindPipeline(ctx, pl.ID)
		assert.NoError(t, err)
		assert.Equal(t, expection.status, pl2.Props.Status)
		assert.Equal(t, expection.errors, pl2.Props.Errors)
	}

}
