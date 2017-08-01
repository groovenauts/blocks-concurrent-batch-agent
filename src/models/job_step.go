package models

import (
	"fmt"
)

type JobStepStatus int

const (
	UnknownJobStepStatus JobStepStatus = iota
	STARTING
	SUCCESS
	FAILURE
)

var JobStepStatusFromString = map[string]JobStepStatus{
	"STARTING": STARTING,
	"SUCCESS":  SUCCESS,
	"FAILURE":  FAILURE,
}

var JobStepStatusToString = map[JobStepStatus]string{
	STARTING: "STARTING",
	SUCCESS:  "SUCCESS",
	FAILURE:  "FAILURE",
}

func (jss JobStepStatus) String() string {
	r, ok := JobStepStatusToString[jss]
	if !ok {
		return "<Invalid JobStepStatus>"
	}
	return r
}

func ParseJobStepStatus(s string) (JobStepStatus, error) {
	jss, ok := JobStepStatusFromString[s]
	if ok {
		return jss, nil
	}
	return UnknownJobStepStatus, fmt.Errorf("Unknown JobStepStatus %q", s)
}

type JobStep int

const (
	UnknownJobStep JobStep = iota
	INITIALIZING
	DOWNLOADING
	EXECUTING
	UPLOADING
	CLEANUP
	NACKSENDING
	CANCELLING
	ACKSENDING
)

var JobStepFromString = map[string]JobStep{
	"INITIALIZING": INITIALIZING,
	"DOWNLOADING":  DOWNLOADING,
	"EXECUTING":    EXECUTING,
	"UPLOADING":    UPLOADING,
	"CLEANUP":      CLEANUP,
	"NACKSENDING":  NACKSENDING,
	"CANCELLING":   CANCELLING,
	"ACKSENDING":   ACKSENDING,
}

var JobStepToString = map[JobStep]string{
	INITIALIZING: "INITIALIZING",
	DOWNLOADING:  "DOWNLOADING",
	EXECUTING:    "EXECUTING",
	UPLOADING:    "UPLOADING",
	CLEANUP:      "CLEANUP",
	NACKSENDING:  "NACKSENDING",
	CANCELLING:   "CANCELLING",
	ACKSENDING:   "ACKSENDING",
}

func (js JobStep) String() string {
	r, ok := JobStepToString[js]
	if !ok {
		return "<Invalid JobStep>"
	}
	return r
}

func ParseJobStep(s string) (JobStep, error) {
	js, ok := JobStepFromString[s]
	if ok {
		return js, nil
	}
	return UnknownJobStep, fmt.Errorf("Unknown JobStep %q", s)
}
