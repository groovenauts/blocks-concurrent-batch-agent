package model

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Scaler struct {
	igServicer InstanceGroupServicer
}

func NewScaler(ctx context.Context) (*Scaler, error) {
	igServicer, err := DefaultInstanceGroupServicer(ctx)
	if err != nil {
		return nil, err
	}
	return &Scaler{igServicer: igServicer}, nil
}

func WithScaler(ctx context.Context, f func(*Scaler) error) error {
	scaler, err := NewScaler(ctx)
	if err != nil {
		return err
	}
	return f(scaler)
}

func (s *Scaler) Process(ctx context.Context, pl *InstanceGroup) (*CloudAsyncOperation, error) {
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
