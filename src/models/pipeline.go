package models

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	pubsub "google.golang.org/api/pubsub/v1"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

// Status constants
type Status int

const (
	Uninitialized Status = iota
	Broken
	Pending
	Waiting
	Reserved
	Building
	Deploying
	Opened
	Closing
	ClosingError
	Closed
)

var StatusStrings = map[Status]string{
	Uninitialized: "uninitialized",
	Broken:        "broken",
	Pending:       "pending",  // Go Waiting when all of the dependencies are satisfied
	Waiting:       "waiting",  // Go Reserved when the pipeline has enough tokens for this TokenConsumption
	Reserved:      "reserved", // Go Building when the pipeline is being built
	Building:      "building",
	Deploying:     "deploying",
	Opened:        "opened",
	Closing:       "closing",
	ClosingError:  "closing_error",
	Closed:        "closed",
}

func (st Status) String() string {
	res, ok := StatusStrings[st]
	if !ok {
		return "Invalid Status: " + strconv.Itoa(int(st))
	}
	return res
}

type (
	// See https://godoc.org/google.golang.org/api/deploymentmanager/v2#OperationErrorErrors
	DeploymentError struct {
		// Code: [Output Only] The error type identifier for this error.
		Code string `json:"code,omitempty"`

		// Location: [Output Only] Indicates the field in the request that
		// caused the error. This property is optional.
		Location string `json:"location,omitempty"`

		// Message: [Output Only] An optional, human-readable error message.
		Message string `json:"message,omitempty"`
	}

	PipelineVmDisk struct {
		// DiskName    string `json:"disk_name,omitempty"` // Don't support diskName to keep simple naming rule
		DiskSizeGb  int    `json:"disk_size_gb,omitempty"`
		DiskType    string `json:"disk_type,omitempty"`
		SourceImage string `json:"source_image" validate:"required"`
	}

	ActionLog struct {
		Time time.Time
		Name string
	}

	Pipeline struct {
		ID                     string            `json:"id"             datastore:"-"`
		Organization           *Organization     `json:"-"              validate:"required" datastore:"-"`
		Name                   string            `json:"name"           validate:"required"`
		ProjectID              string            `json:"project_id"     validate:"required"`
		Zone                   string            `json:"zone"           validate:"required"`
		BootDisk               PipelineVmDisk    `json:"boot_disk"`
		MachineType            string            `json:"machine_type"   validate:"required"`
		Preemptible            bool              `json:"preemptible,omitempty"`
		StackdriverAgent       bool              `json:"stackdriver_agent,omitempty"`
		TargetSize             int               `json:"target_size"    validate:"required"`
		ContainerSize          int               `json:"container_size" validate:"required"`
		ContainerName          string            `json:"container_name" validate:"required"`
		Command                string            `json:"command"` // allow blank
		Status                 Status            `json:"status"`
		Dryrun                 bool              `json:"dryrun"`
		DeploymentName         string            `json:"deployment_name"`
		DeployingOperationName string            `json:"deploying_operation_name"`
		DeployingErrors        []DeploymentError `json:"deploying_errors"`
		ClosingOperationName   string            `json:"closing_operation_name"`
		ClosingErrors          []DeploymentError `json:"closing_errors"`
		TokenConsumption       int               `json:"token_consumption"`
		Dependency             Dependency        `json:"dependency,omitempty"`
		ClosePolicy            ClosePolicy       `json:"close_policy,omitempty"`
		CreatedAt              time.Time         `json:"created_at"`
		UpdatedAt              time.Time         `json:"updated_at"`
		ActionLogs             []ActionLog       `json:"action_logs"`
	}
)

func (m *Pipeline) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *Pipeline) Create(ctx context.Context) error {
	return m.CreateWith(ctx, m.PutWithNewKey)
}

func (m *Pipeline) CreateWith(ctx context.Context, f func(ctx context.Context) error) error {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = t
	}

	err := m.Validate()
	if err != nil {
		return err
	}

	return f(ctx)
}

func (m *Pipeline) ReserveOrWait(ctx context.Context, f func(context.Context) error) error {
	log.Debugf(ctx, "Start ReserveOrWait pipeline: %v", m)
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		dep := &m.Dependency
		sat, err := dep.Satisfied(ctx)
		if err != nil {
			return err
		}
		if !sat {
			m.Status = Pending
		} else {
			log.Debugf(ctx, "Dependency is satisfied. pipeline: %v", m)
			org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
			if err != nil {
				return err
			}

			waiting, err := org.PipelineAccessor().WaitingQuery()
			if err != nil {
				return err
			}

			cnt, err := waiting.Count(ctx)
			if err != nil {
				return err
			}

			if cnt > 0 {
				log.Warningf(ctx, "Insufficient tokens; %v has already %v waiting pipelines", org.Name, cnt)
				m.Status = Waiting
			} else {
				newAmount := org.TokenAmount - m.TokenConsumption
				if newAmount < 0 {
					log.Warningf(ctx, "Insufficient tokens; %v has only %v tokens but %v required %v tokens", org.Name, org.TokenAmount, m.Name, m.TokenConsumption)
					m.Status = Waiting
				} else {
					m.Status = Reserved
					org.TokenAmount = newAmount
					err = org.Update(ctx)
					if err != nil {
						return err
					}
				}
			}
		}

		return f(ctx)
	}, GetTransactionOptions(ctx))

	if err != nil {
		log.Errorf(ctx, "Transaction failed: %v\n", err)
		return err
	}

	return nil
}

func (m *Pipeline) CreateWithReserveOrWait(ctx context.Context) error {
	return m.CreateWith(ctx, func(ctx context.Context) error {
		return m.ReserveOrWait(ctx, func(ctx context.Context) error {
			return m.PutWithNewKey(ctx)
		})
	})
}

func (m *Pipeline) UpdateIfReserveOrWait(ctx context.Context) error {
	original := m.Status
	err := m.ReserveOrWait(ctx, func(ctx context.Context) error {
		if original == m.Status {
			return nil
		}
		return m.Update(ctx)
	})
	return err
}

func (m *Pipeline) PutWithNewKey(ctx context.Context) error {
	parentKey, err := datastore.DecodeKey(m.Organization.ID)
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(ctx, "Pipelines", parentKey)
	if err != nil {
		return err
	}

	res, err := datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	m.ID = res.Encode()

	return nil
}

func (m *Pipeline) Destroy(ctx context.Context) error {
	if m.Status != Closed {
		return &InvalidOperation{
			Msg: fmt.Sprintf("Can't destroy pipeline which is %v. Close before delete.", m.Status),
		}
	}
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	if err := datastore.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) Update(ctx context.Context) error {
	if m.Organization == nil {
		err := m.LoadOrganization(ctx)
		if err != nil {
			return err
		}
	}

	m.UpdatedAt = time.Now()
	if m.Organization == nil {
		err := m.LoadOrganization(ctx)
		if err != nil {
			return err
		}
	}

	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) RefreshHandler(ctx context.Context) func(*[]DeploymentError) error {
	return m.RefreshHandlerWith(ctx, nil)
}

func (m *Pipeline) RefreshHandlerWith(ctx context.Context, pipelineProcesser func(*Pipeline) error) func(*[]DeploymentError) error {
	return func(errors *[]DeploymentError) error {
		switch m.Status {
		case Deploying:
			if errors != nil {
				return m.FailDeploying(ctx, errors)
			} else {
				return m.CompleteDeploying(ctx)
			}
		case Closing:
			if errors != nil {
				return m.FailClosing(ctx, errors)
			} else {
				return m.CompleteClosing(ctx, pipelineProcesser)
			}
		default:
			return &InvalidOperation{Msg: fmt.Sprintf("Invalid Status %v to handle refreshing Pipline %q\n", m.Status, m.ID)}
		}
	}
}

func (m *Pipeline) StateTransition(ctx context.Context, froms []Status, to Status) error {
	allowed := false
	for _, from := range froms {
		if m.Status == from {
			allowed = true
			break
		}
	}
	if !allowed {
		return &InvalidStateTransition{fmt.Sprintf("Forbidden state transition from %v to %v for pipeline: %v", m.Status, to, m)}
	}
	m.Status = to
	return m.Update(ctx)
}

func (m *Pipeline) StartBuilding(ctx context.Context) error {
	m.AddActionLog(ctx, "build-started")
	return m.StateTransition(ctx, []Status{Reserved, Building}, Building)
}

func (m *Pipeline) FinishBuilding(ctx context.Context) {
}

func (m *Pipeline) StartDeploying(ctx context.Context, deploymentName, operationName string) error {
	m.DeploymentName = deploymentName
	m.DeployingOperationName = operationName
	return m.StateTransition(ctx, []Status{Building}, Deploying)
}

func (m *Pipeline) FailDeploying(ctx context.Context, errors *[]DeploymentError) error {
	m.AddActionLog(ctx, "build-finished")
	m.DeployingErrors = *errors
	return m.StateTransition(ctx, []Status{Deploying}, Broken)
}

func (m *Pipeline) CompleteDeploying(ctx context.Context) error {
	m.AddActionLog(ctx, "build-finished")
	return m.StateTransition(ctx, []Status{Deploying}, Opened)
}

func (m *Pipeline) StartClosing(ctx context.Context, operationName string) error {
	m.AddActionLog(ctx, "close-started")
	m.ClosingOperationName = operationName
	return m.StateTransition(ctx, []Status{Opened, Closing}, Closing)
}

func (m *Pipeline) FailClosing(ctx context.Context, errors *[]DeploymentError) error {
	m.AddActionLog(ctx, "close-finished")
	m.ClosingErrors = *errors
	return m.StateTransition(ctx, []Status{Closing}, ClosingError)
}

func (m *Pipeline) CompleteClosing(ctx context.Context, pipelineProcesser func(*Pipeline) error) error {
	m.AddActionLog(ctx, "close-finished")
	m.Update(ctx)
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
		if err != nil {
			return err
		}
		newTokenAmount := org.TokenAmount + m.TokenConsumption

		waitings, err := org.PipelineAccessor().GetWaitings(ctx)
		if err != nil {
			return err
		}

		for _, waiting := range waitings {
			if newTokenAmount < waiting.TokenConsumption {
				break
			}
			newTokenAmount = newTokenAmount - waiting.TokenConsumption
			waiting.Status = Reserved
			err := waiting.Update(ctx)
			if err != nil {
				return err
			}
			if pipelineProcesser != nil {
				err := pipelineProcesser(waiting)
				if err != nil {
					return err
				}
			}
		}

		org.TokenAmount = newTokenAmount
		err = org.Update(ctx)
		if err != nil {
			return err
		}

		return m.StateTransition(ctx, []Status{Closing}, Closed)
	}, GetTransactionOptions(ctx))
}

func (m *Pipeline) LoadOrganization(ctx context.Context) error {
	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to decode Key of pipeline %v because of %v\n", m.ID, err)
		return err
	}
	orgKey := key.Parent()
	if orgKey == nil {
		log.Errorf(ctx, "Pipline key has no parent. ID: %v\n", m.ID)
		panic("Invalid pipeline key")
	}
	org, err := GlobalOrganizationAccessor.FindByKey(ctx, orgKey)
	if err != nil {
		return err
	}
	m.Organization = org
	return nil
}

func (m *Pipeline) JobAccessor() *JobAccessor {
	return &JobAccessor{Parent: m}
}

func (m *Pipeline) Reload(ctx context.Context) error {
	err := GlobalPipelineAccessor.LoadByID(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) PublishJobs(ctx context.Context) error {
	m.AddActionLog(ctx, "publish-started")
	err := m.Update(ctx)
	if err != nil {
		return err
	}
	defer func() {
		m.AddActionLog(ctx, "publish-finished")
		m.Update(ctx)
		// ignore error
	}()

	accessor := m.JobAccessor()
	jobs, err := accessor.All(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if job.Status == Ready {
			job.Pipeline = m
			_, err := job.Publish(ctx)
			if err != nil {
				log.Errorf(ctx, "Failed to publish job %v because of %v\n", job, err)
				return err
			}
		}
	}

	return nil
}

func (m *Pipeline) JobTopicName() string {
	return fmt.Sprintf("%s-job-topic", m.Name)
}

func (m *Pipeline) JobTopicFqn() string {
	return fmt.Sprintf("projects/%s/topics/%s", m.ProjectID, m.JobTopicName())
}

func (m *Pipeline) ProgressSubscriptionName() string {
	return fmt.Sprintf("%s-progress-subscription", m.Name)
}

func (m *Pipeline) ProgressSubscriptionFqn() string {
	return fmt.Sprintf("projects/%s/subscriptions/%s", m.ProjectID, m.ProgressSubscriptionName())
}

func (m *Pipeline) IDHex() string {
	return hex.EncodeToString([]byte(m.ID))
}

func (m *Pipeline) PullAndUpdateJobStatus(ctx context.Context) error {
	s := &PubsubSubscriber{MessagePerPull: 10}
	err := s.setup(ctx)
	if err != nil {
		return err
	}

	accessor := m.JobAccessor()
	err = s.subscribe(ctx, m.ProgressSubscriptionFqn(), func(recvMsg *pubsub.ReceivedMessage) error {
		attrs := recvMsg.Message.Attributes
		jobId := attrs[JobIdKey]
		job, err := accessor.Find(ctx, jobId)
		if err != nil {
			return err
		}
		step, err := ParseJobStep(attrs["step"])
		if err != nil {
			return err
		}
		stepStatus, err := ParseJobStepStatus(attrs["step_status"])
		if err != nil {
			return err
		}
		completed, err := strconv.ParseBool(attrs["completed"])
		if err != nil {
			return err
		}
		job.Hostname = m.stringFromMapWithDefault(attrs, "host", "unknown")
		job.Zone = m.stringFromMapWithDefault(attrs, "zone", "unknown")
		job.StartTime = m.stringFromMapWithDefault(attrs, "job.start-time", "")
		job.FinishTime = m.stringFromMapWithDefault(attrs, "job.finish-time", "")
		err = job.UpdateStatusIfGreaterThanBefore(ctx, completed, step, stepStatus)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) stringFromMapWithDefault(src map[string]string, key, defaultValue string) string {
	r, ok := src[key]
	if !ok {
		return defaultValue
	}
	return r
}

func (m *Pipeline) AddActionLog(ctx context.Context, name string) {
	if m.ActionLogs == nil {
		m.ActionLogs = []ActionLog{}
	}
	m.ActionLogs = append(m.ActionLogs, ActionLog{
		Time: time.Now(),
		Name: name,
	})
}
