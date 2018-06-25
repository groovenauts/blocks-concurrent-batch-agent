package model

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/goadesign/goa/uuid"

	"github.com/mjibson/goon"
)

type PipelineBaseStatus string

const (
	Opening               PipelineBaseStatus = "opening"
	OpeningError          PipelineBaseStatus = "opening_error"
	Hibernating           PipelineBaseStatus = "hibernating"
	Waking                PipelineBaseStatus = "waking"
	WakingError           PipelineBaseStatus = "waking_error"
	Awake                 PipelineBaseStatus = "awake"
	HibernationChecking   PipelineBaseStatus = "hibernation_checking"
	HibernationGoing      PipelineBaseStatus = "hibernation_going"
	HibernationGoingError PipelineBaseStatus = "hibernation_going_error"
	Closing               PipelineBaseStatus = "closing"
	ClosingError          PipelineBaseStatus = "closing_error"
	Closed                PipelineBaseStatus = "closed"
)

type PipelineContainer struct {
	name              string
	size              int
	command           string
	options           string
	stackdriver_agent boolean
}

type PipelineBase struct {
	Name             string
	InstanceGroup    InstanceGroupBody
	Container        PipelineContainer
	HibernationDelay int
	Status           PipelineBaseStatus
	IntanceGroupID   string
}
