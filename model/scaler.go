package models

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

func (s *Scaler) Process(ctx context.Context, pl *Pipeline) (*PipelineOperation, error) {
	if !pl.CanScale() {
		log.Infof(ctx, "Quit Scaler#Process because the pipeline can't scale because of %v\n", pl.JobScaler)
		return nil, nil
	}
	workingJobCount, err := pl.JobAccessor().WorkingCount(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to get workingJobCount of %v because of %v\n", pl.ID, err)
		return nil, err
	}
	capacity := pl.InstanceSize * pl.ContainerSize
	shortage := workingJobCount - capacity
	if shortage < 1 {
		log.Debugf(ctx, "Pipeline has enough %d instances for %d jobs\n", pl.InstanceSize, workingJobCount)
		return nil, err
	}
	newInstanceSize := workingJobCount / pl.ContainerSize
	if m := workingJobCount % pl.ContainerSize; m > 0 {
		newInstanceSize += 1
	}

	if newInstanceSize > pl.JobScaler.MaxInstanceSize {
		if pl.JobScaler.MaxInstanceSize > pl.InstanceSize {
			log.Warningf(ctx, "Can't assign %d VMs but can assign %d VMs as max\n", newInstanceSize, pl.JobScaler.MaxInstanceSize)
			newInstanceSize = pl.JobScaler.MaxInstanceSize
		} else {
			log.Warningf(ctx, "Quit increacing instances to %d because of MaxInstanceSize %d\n", newInstanceSize, pl.JobScaler.MaxInstanceSize)
			return nil, nil
		}
	}

	ope, err := s.igServicer.Resize(pl.ProjectID, pl.Zone, pl.DeploymentName+"-igm", int64(newInstanceSize))
	if err != nil {
		log.Errorf(ctx, "Failed to Resize %v/%v/%v to %d\n", pl.ProjectID, pl.Zone, pl.DeploymentName, newInstanceSize)
		return nil, err
	}

	operation := &PipelineOperation{
		Pipeline:      pl,
		ProjectID:     pl.ProjectID,
		Zone:          pl.Zone,
		Service:       "compute",
		Name:          ope.Name,
		OperationType: ope.OperationType,
		Status:        ope.Status,
		Logs: []OperationLog{
			OperationLog{CreatedAt: time.Now(), Message: "Start"},
		},
	}
	err = operation.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create PipelineOperation: %v because of %v\n", operation, err)
		return nil, err
	}

	pl.InstanceSize = newInstanceSize
	err = pl.Update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Pipeline InstanceSize : %v because of %v\n", pl, err)
		return nil, err
	}

	return operation, nil
}
