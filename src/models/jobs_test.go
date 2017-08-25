package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobsAllFinished(t *testing.T) {
	succeeded := func(job *Job) bool { return job.Status == Success }

	type Pattern struct {
		f  bool
		a  bool
		st []JobStatus
	}

	patterns := []Pattern{
		Pattern{f: false, a: true, st: []JobStatus{}},
		Pattern{f: false, a: false, st: []JobStatus{Executing}},
		Pattern{f: true, a: false, st: []JobStatus{Failure}},
		Pattern{f: true, a: true, st: []JobStatus{Success}},
		Pattern{f: false, a: false, st: []JobStatus{Executing, Executing}},
		Pattern{f: false, a: false, st: []JobStatus{Executing, Failure}},
		Pattern{f: false, a: false, st: []JobStatus{Executing, Success}},
		Pattern{f: false, a: false, st: []JobStatus{Failure, Executing}},
		Pattern{f: true, a: false, st: []JobStatus{Failure, Failure}},
		Pattern{f: true, a: false, st: []JobStatus{Failure, Success}},
		Pattern{f: false, a: false, st: []JobStatus{Success, Executing}},
		Pattern{f: true, a: false, st: []JobStatus{Success, Failure}},
		Pattern{f: true, a: true, st: []JobStatus{Success, Success}},
	}

	boolToAssert := func(b bool) func(assert.TestingT, bool, ...interface{}) bool {
		if b {
			return assert.True
		} else {
			return assert.False
		}
	}

	for _, ptn := range patterns {
		jobs := Jobs{}
		for _, st := range ptn.st {
			jobs = append(jobs, &Job{Status: st})
		}
		ff := boolToAssert(ptn.f)
		af := boolToAssert(ptn.a)

		if !ff(t, jobs.AllFinished()) {
			fmt.Printf("ff failure pattern: %v\n", ptn)
		}
		if !af(t, jobs.All(succeeded)) {
			fmt.Printf("af failure pattern: %v\n", ptn)
		}
	}
}
