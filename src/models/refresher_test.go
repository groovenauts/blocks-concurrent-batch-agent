package models

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
	ope, _ := d.GetOperation(ctx, project, "")
	return &deploymentmanager.Deployment{Operation: ope}, nil
}
func (d *TestDeployerRunning) GetOperation(ctx context.Context, project string, operation string) (*deploymentmanager.Operation, error) {
	return &deploymentmanager.Operation{Status: "RUNNING"}, nil
}

type TestDeployerOK struct{}

func (d *TestDeployerOK) Delete(ctx context.Context, project string, deployment string) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerOK) Insert(ctx context.Context, project string, deployment *deploymentmanager.Deployment) (*deploymentmanager.Operation, error) {
	return nil, nil
}
func (d *TestDeployerOK) Get(ctx context.Context, project string, deployment string) (*deploymentmanager.Deployment, error) {
	ope, _ := d.GetOperation(ctx, project, "")
	return &deploymentmanager.Deployment{Operation: ope}, nil
}
func (d *TestDeployerOK) GetOperation(ctx context.Context, project string, operation string) (*deploymentmanager.Operation, error) {
	return &deploymentmanager.Operation{
		Status: "DONE",
		Error:  nil,
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
	ope, _ := d.GetOperation(ctx, project, "")
	return &deploymentmanager.Deployment{Operation: ope}, nil
}
func (d *TestDeployerError) GetOperation(ctx context.Context, project string, operation string) (*deploymentmanager.Operation, error) {
	return &deploymentmanager.Operation{
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
	}, nil
}

func TestRefresherProcessForDeploying(t *testing.T) {
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
			status:   Deploying,
			deployer: &TestDeployerRunning{},
			errors:   nil,
		},
		Expection{
			status:   Opened,
			deployer: &TestDeployerOK{},
			errors:   nil,
		},
		Expection{
			status:   Broken,
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

	org1 := &Organization{
		Name: "org01",
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	for _, expection := range expections {
		pl := Pipeline{
			Organization: org1,
			Name:         "pipeline01",
			ProjectID:    proj,
			Zone:         "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:    "f1-micro",
			TargetSize:     2,
			ContainerSize:  2,
			ContainerName:  "groovenauts/batch_type_iot_example:0.3.1",
			Command:        "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
			DeploymentName: "pipeline01",
			Status:         Deploying,
		}
		err = pl.Create(ctx)

		r := &Refresher{deployer: expection.deployer}
		err = r.Process(ctx, &pl)
		assert.NoError(t, err)
		pl2, err := GlobalPipelineAccessor.Find(ctx, pl.ID)
		assert.NoError(t, err)
		assert.Equal(t, expection.status, pl2.Status)
		assert.Equal(t, expection.errors, pl2.DeployingErrors)
	}

}

func TestRefresherProcessForClosing(t *testing.T) {
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
			status:   Closing,
			deployer: &TestDeployerRunning{},
			errors:   nil,
		},
		Expection{
			status:   Closed,
			deployer: &TestDeployerOK{},
			errors:   nil,
		},
		Expection{
			status:   Closing_error,
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

	org1 := &Organization{
		Name: "org01",
		TokenAmount: 10,
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	for _, expection := range expections {
		pl := Pipeline{
			Organization: org1,
			Name:         "pipeline01",
			ProjectID:    proj,
			Zone:         "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:    "f1-micro",
			TargetSize:     2,
			ContainerSize:  2,
			ContainerName:  "groovenauts/batch_type_iot_example:0.3.1",
			Command:        "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
			DeploymentName: "pipeline01",
			Status:         Closing,
			TokenConsumption: 0,
		}
		err = pl.Create(ctx)
		assert.NoError(t, err)

		pl.TokenConsumption = 2
		err = pl.Update(ctx)
		assert.NoError(t, err)

		orgBefore, err := GlobalOrganizationAccessor.Find(ctx, org1.ID)
		assert.NoError(t, err)

		r := &Refresher{deployer: expection.deployer}
		err = r.Process(ctx, &pl)
		assert.NoError(t, err)
		pl2, err := GlobalPipelineAccessor.Find(ctx, pl.ID)
		assert.NoError(t, err)
		assert.Equal(t, expection.status, pl2.Status)
		assert.Equal(t, expection.errors, pl2.ClosingErrors)

		orgAfter, err := GlobalOrganizationAccessor.Find(ctx, org1.ID)
		assert.NoError(t, err)

		switch expection.status {
		case Closed:
			assert.Equal(t, orgBefore.TokenAmount + pl.TokenConsumption, orgAfter.TokenAmount)
		default:
			assert.Equal(t, orgBefore.TokenAmount, orgAfter.TokenAmount)
		}
	}
}
