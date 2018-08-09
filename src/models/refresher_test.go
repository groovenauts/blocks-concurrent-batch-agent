package models

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/appengine"
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
	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if !assert.NoError(t, err) {
		inst.Close()
		return
	}
	ctx := appengine.NewContext(req)

	type Expection struct {
		status   Status
		deployer DeploymentServicer
		errors   []OperationError
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
			errors: []OperationError{
				OperationError{
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
		pl := &Pipeline{
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
		assert.NoError(t, pl.Create(ctx))

		ope := &PipelineOperation{
			Pipeline:      pl,
			ProjectID:     pl.ProjectID,
			Zone:          pl.Zone,
			Service:       "deploymentmanager",
			Name:          "deployOp",
			OperationType: "insert",
			Status:        "RUNNING",
		}
		assert.NoError(t, ope.Create(ctx))

		updater := &DeploymentUpdater{
			Servicer: expection.deployer,
		}
		assert.NoError(t, ope.ProcessDeploy(ctx, updater))

		pl2, err := GlobalPipelineAccessor.Find(ctx, pl.ID)
		assert.NoError(t, err)
		assert.Equal(t, expection.status, pl2.Status)
		assert.Equal(t, expection.errors, ope.Errors)
	}

}

func TestRefresherProcessForClosing(t *testing.T) {
	opt := &aetest.Options{StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	assert.NoError(t, err)
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if !assert.NoError(t, err) {
		inst.Close()
		return
	}
	ctx := appengine.NewContext(req)

	type Expection struct {
		status   Status
		deployer DeploymentServicer
		errors   []OperationError
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
			status:   ClosingError,
			deployer: &TestDeployerError{},
			errors: []OperationError{
				OperationError{
					Code:     "999",
					Location: "Somewhere",
					Message:  "Something wrong",
				},
			},
		},
	}

	org1 := &Organization{
		Name:        "org01",
		TokenAmount: 10,
	}
	err = org1.Create(ctx)
	assert.NoError(t, err)

	for _, expection := range expections {
		pl := &Pipeline{
			Organization: org1,
			Name:         "pipeline01",
			ProjectID:    proj,
			Zone:         "us-central1-f",
			BootDisk: PipelineVmDisk{
				SourceImage: "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
			},
			MachineType:      "f1-micro",
			TargetSize:       2,
			ContainerSize:    2,
			ContainerName:    "groovenauts/batch_type_iot_example:0.3.1",
			Command:          "bundle exec magellan-gcs-proxy echo %{download_files.0} %{downloads_dir} %{uploads_dir}",
			DeploymentName:   "pipeline01",
			Status:           Closing,
			TokenConsumption: 0,
		}
		assert.NoError(t, pl.Create(ctx))

		// Update TokenConsumption to avoid decreasing organization's token
		pl.TokenConsumption = 2
		assert.NoError(t, pl.Update(ctx))

		ope := &PipelineOperation{
			Pipeline:      pl,
			ProjectID:     pl.ProjectID,
			Zone:          pl.Zone,
			Service:       "deploymentmanager",
			Name:          "closeOp",
			OperationType: "delete",
			Status:        "RUNNING",
		}
		assert.NoError(t, ope.Create(ctx))

		updater := &DeploymentUpdater{
			Servicer: expection.deployer,
		}

		orgBefore, err := GlobalOrganizationAccessor.Find(ctx, org1.ID)
		assert.NoError(t, err)

		called := false
		assert.NoError(t, ope.ProcessClosing(ctx, updater, func(*Pipeline) error {
			called = true
			return nil
		}))

		pl2, err := GlobalPipelineAccessor.Find(ctx, pl.ID)
		assert.NoError(t, err)
		assert.Equal(t, expection.status, pl2.Status)
		assert.Equal(t, expection.errors, ope.Errors)

		orgAfter, err := GlobalOrganizationAccessor.Find(ctx, org1.ID)
		assert.NoError(t, err)

		switch expection.status {
		case Closed:
			assert.Equal(t, orgBefore.TokenAmount+pl.TokenConsumption, orgAfter.TokenAmount)
			assert.False(t, called) // Handler isn't called because there are no waiting pipeline
		case Closing, ClosingError:
			assert.Equal(t, orgBefore.TokenAmount, orgAfter.TokenAmount)
			assert.False(t, called)
		default:
			assert.Fail(t, "Unexpected status: %v exception: %v\n", expection.status, expection)
		}
	}
}
