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

type PipelineStatus string

const (
	CurrentPreparing      PipelineStatus = "current_preparing"
	CurrentPreparingError PipelineStatus = "current_preparing_error"
	Running               PipelineStatus = "running"
	NextPreparing         PipelineStatus = "next_preparing"
	Stopping              PipelineStatus = "stopping"
	StoppingError         PipelineStatus = "stopping_error"
	Stopped               PipelineStatus = "stopped"
)

type Pipeline struct {
	Name             string
	InstanceGroup    InstanceGroupBody
	Container        PipelineContainer
	HibernationDelay int
	Status           PipelineStatus
	IntanceGroupID   string
}
