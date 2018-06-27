package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type InstanceGroupScaler struct {
	igServicer InstanceGroupServicer
}

func NewInstanceGroupScaler(ctx context.Context) (*InstanceGroupScaler, error) {
	igServicer, err := DefaultInstanceGroupServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &InstanceGroupScaler{igServicer: igServicer}, nil
}

func WithInstanceGroupScaler(ctx context.Context, f func(*InstanceGroupScaler) error) error {
	scaler, err := NewInstanceGroupScaler(ctx)
	if err != nil {
		return err
	}
	return f(scaler)
}

func (s *InstanceGroupScaler) Process(ctx context.Context, pl *InstanceGroup) (*CloudAsyncOperation, error) {
	ope, err := s.igServicer.Resize(pl.ProjectID, pl.Zone, pl.DeploymentName+"-igm", int64(pl.InstanceSizeRequested))
	if err != nil {
		log.Errorf(ctx, "Failed to Resize %v/%v/%v to %d\n", pl.ProjectID, pl.Zone, pl.DeploymentName, pl.InstanceSizeRequested)
		return nil, err
	}

	operation := &CloudAsyncOperation{
		OwnerType:     "InstanceGroup",
		OwnerID:       pl.Id,
		ProjectId:     pl.ProjectID,
		Zone:          pl.Zone,
		Service:       "compute",
		Name:          ope.Name,
		OperationType: ope.OperationType,
		Status:        ope.Status,
		Logs: []CloudAsyncOperationLog{
			CloudAsyncOperationLog{CreatedAt: time.Now(), Message: "Start"},
		},
	}

	return operation, nil
}
