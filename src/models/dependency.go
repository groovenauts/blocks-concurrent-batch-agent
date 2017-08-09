package models

import (
	"golang.org/x/net/context"
)

type DependencyCondition int

const (
	OnSuccess DependencyCondition = iota
	OnFailure
	OnFinish // On Failure or Success
)

type Dependency struct {
	Condition DependencyCondition `json:"condition"`
	JobIDs    []string            `json:"job_ids"`
}

func (m *Dependency) Satisfied(ctx context.Context) (bool, error) {
	for _, jobId := range m.JobIDs {
		job, err := GlobalJobAccessor.Find(ctx, jobId)
		if err != nil {
			return false, err
		}
		var r bool
		switch m.Condition {
		case OnFailure:
			r = job.Status == Failure
		case OnSuccess:
			r = job.Status == Success
		case OnFinish:
			r = job.Status.IsFinished()
		default:
			r = false
		}
		if !r {
			return false, nil
		}
	}
	return true, nil
}
