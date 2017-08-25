package models

import (
	"strconv"
)

type ClosePolicy int

const (
	CloseAnyway ClosePolicy = iota
	CloseOnAllSuccess
	CloseNever
)

var ClosePolicyStrings = map[ClosePolicy]string{
	CloseAnyway:       "CloseAnyway",
	CloseOnAllSuccess: "CloseOnAllSuccess",
	CloseNever:        "CloseNever",
}

func (cp ClosePolicy) String() string {
	res, ok := ClosePolicyStrings[cp]
	if !ok {
		return "Invalid ClosePolicy: " + strconv.Itoa(int(cp))
	}
	return res
}

func (cp ClosePolicy) Match(jobs Jobs) bool {
	switch cp {
	case CloseOnAllSuccess:
		return jobs.All(func(job *Job) bool {
			return job.Status == Success
		})
	case CloseNever:
		return false
	default:
		return true
	}
}
