package models

import (
	"encoding/hex"
	"fmt"
	"regexp"
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
	HibernationChecking
	HibernationStarting
	HibernationProcessing
	HibernationError
	Hibernating
	Closing
	ClosingError
	Closed
)

var StatusStrings = map[Status]string{
	Uninitialized:         "uninitialized",
	Broken:                "broken",
	Pending:               "pending",  // Go Waiting when all of the dependencies are satisfied
	Waiting:               "waiting",  // Go Reserved when the pipeline has enough tokens for this TokenConsumption
	Reserved:              "reserved", // Go Building when the pipeline is being built
	Building:              "building",
	Deploying:             "deploying",
	Opened:                "opened",
	HibernationChecking:   "hibernation_waiting",
	HibernationStarting:   "hibernation_starting",
	HibernationProcessing: "hibernation_processing",
	HibernationError:      "hibernation_error",
	Hibernating:           "hibernating",
	Closing:               "closing",
	ClosingError:          "closing_error",
	Closed:                "closed",
}

func (st Status) String() string {
	res, ok := StatusStrings[st]
	if !ok {
		return "Invalid Status: " + strconv.Itoa(int(st))
	}
	return res
}

type Statuses []Status

var (
	StatusesNotDeployedYet         = Statuses{Uninitialized, Pending, Waiting, Reserved}
	StatusesNowDeploying           = Statuses{Building, Deploying}
	StatusesOpened                 = Statuses{Opened, HibernationChecking}
	StatusesAlreadyClosing         = Statuses{Closing, ClosingError, Closed}
	StatusesHibernationInProgresss = Statuses{HibernationStarting, HibernationProcessing, HibernationError}
	StatusesHibernating            = Statuses{Hibernating}
)

func (sts Statuses) Include(t Status) bool {
	for _, st := range sts {
		if st == t {
			return true
		}
	}
	return false
}

type (
	PipelineVmDisk struct {
		// DiskName    string `json:"disk_name,omitempty"` // Don't support diskName to keep simple naming rule
		DiskSizeGb  int    `json:"disk_size_gb,omitempty"`
		DiskType    string `json:"disk_type,omitempty"`
		SourceImage string `json:"source_image" validate:"required"`
	}

	Accelerators struct {
		Count int    `json:"count"`
		Type  string `json:"type"`
	}

	Pipeline struct {
		ID                   string         `json:"id"             datastore:"-"`
		Organization         *Organization  `json:"-"              validate:"required" datastore:"-"`
		Name                 string         `json:"name"           validate:"required"`
		ProjectID            string         `json:"project_id"     validate:"required"`
		Zone                 string         `json:"zone"           validate:"required"`
		BootDisk             PipelineVmDisk `json:"boot_disk"`
		MachineType          string         `json:"machine_type"   validate:"required"`
		GpuAccelerators      Accelerators   `json:"gpu_accelerators,omitempty"`
		Preemptible          bool           `json:"preemptible,omitempty"`
		StackdriverAgent     bool           `json:"stackdriver_agent,omitempty"`
		TargetSize           int            `json:"target_size"    validate:"required"`
		ContainerSize        int            `json:"container_size" validate:"required"`
		ContainerName        string         `json:"container_name" validate:"required"`
		Command              string         `json:"command"` // allow blank
		Status               Status         `json:"status"`
		Cancelled            bool           `json:"cancelled"`
		Dryrun               bool           `json:"dryrun"`
		DeploymentName       string         `json:"deployment_name"`
		TokenConsumption     int            `json:"token_consumption"`
		Dependency           Dependency     `json:"dependency,omitempty"`
		ClosePolicy          ClosePolicy    `json:"close_policy,omitempty"`
		HibernationDelay     int            `json:"hibernation_delay,omitempty"` // seconds
		HibernationStartedAt time.Time      `json:"hibernation_started_at,omitempty"`
		CreatedAt            time.Time      `json:"created_at"`
		UpdatedAt            time.Time      `json:"updated_at"`
	}
)

const (
	Ubuntu16ImageFamily = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/family/ubuntu-1604-lts"
)

var ImagesRecommendedForGpu = []string{
	Ubuntu16ImageFamily,
}
var Ubuntu1604Regexp = regexp.MustCompile(`ubuntu.*1604`)

func PipelineStructLevelValidation(sl validator.StructLevel) {
	pl := sl.Current().Interface().(Pipeline)
	bd := pl.BootDisk
	if pl.GpuAccelerators.Count > 0 {
		if !Ubuntu1604Regexp.MatchString(bd.SourceImage) {
			sl.ReportError(bd.SourceImage, "SourceImage", "", "source_image", "Invalid Image for GPU")
		}
	}
}

func (m *Pipeline) Validate() error {
	validator := validator.New()
	validator.RegisterStructValidation(PipelineStructLevelValidation, Pipeline{})
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
	return m.StateTransition(ctx, []Status{Reserved, Building}, Building)
}

func (m *Pipeline) StartDeploying(ctx context.Context, deploymentName string) error {
	m.DeploymentName = deploymentName
	return m.StateTransition(ctx, []Status{Building}, Deploying)
}

func (m *Pipeline) FailDeploying(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{Deploying}, Broken)
}

func (m *Pipeline) CompleteDeploying(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{Deploying}, Opened)
}

func (m *Pipeline) WaitHibernation(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{Opened}, HibernationChecking)
}

func (m *Pipeline) StartHibernation(ctx context.Context) error {
	m.HibernationStartedAt = time.Now()
	return m.StateTransition(ctx, []Status{HibernationChecking}, HibernationStarting)
}

func (m *Pipeline) ProcessHibernation(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{HibernationStarting, HibernationProcessing}, HibernationProcessing)
}

func (m *Pipeline) FailHibernation(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{HibernationProcessing}, HibernationError)
}

func (m *Pipeline) CompleteHibernation(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{HibernationProcessing}, Hibernating)
}

func (m *Pipeline) BackToBeOpened(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{HibernationChecking}, Opened)
}

func (m *Pipeline) BackToBeReserved(ctx context.Context) error {
	m.HibernationStartedAt = time.Time{}
	return m.StateTransition(ctx, []Status{Hibernating}, Reserved)
}

func (m *Pipeline) StartClosing(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{Opened, Closing}, Closing)
}

func (m *Pipeline) FailClosing(ctx context.Context) error {
	return m.StateTransition(ctx, []Status{Closing}, ClosingError)
}

func (m *Pipeline) CompleteClosing(ctx context.Context, pipelineProcesser func(*Pipeline) error) error {
	err := m.LoadOrganization(ctx)
	if err != nil {
		return err
	}
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := m.CancelLivingJobs(ctx)
		if err != nil {
			return err
		}

		org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
		if err != nil {
			return err
		}

		org.GetBackToken(ctx, m, func() error {
			return org.StartWaitingPipelines(ctx, pipelineProcesser)
		})

		return m.StateTransition(ctx, []Status{Closing}, Closed)
	}, GetTransactionOptions(ctx))
}

func (m *Pipeline) CloseIfHibernating(ctx context.Context) error {
	return m.StateTransition(ctx, StatusesHibernating, ClosingError)
}

func (m *Pipeline) CancelLivingJobs(ctx context.Context) error {
	accessor := m.JobAccessor()
	jobs, err := accessor.All(ctx)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Status.Living() {
			err = job.Cancel(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Pipeline) Cancel(ctx context.Context) error {
	m.Cancelled = true
	// m.AddActionLog(ctx, "cancelled")
	switch {
	case StatusesNotDeployedYet.Include(m.Status) || StatusesHibernating.Include(m.Status):
		m.Status = Closed
	}
	return m.Update(ctx)
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

func (m *Pipeline) OperationAccessor() *PipelineOperationAccessor {
	return &PipelineOperationAccessor{Parent: m}
}

func (m *Pipeline) Reload(ctx context.Context) error {
	err := GlobalPipelineAccessor.LoadByID(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) PublishJobs(ctx context.Context) error {
	// m.AddActionLog(ctx, "publish-started")
	err := m.Update(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// m.AddActionLog(ctx, "publish-finished")
		m.Update(ctx)
		// ignore error
	}()

	accessor := m.JobAccessor()
	jobs, err := accessor.All(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		switch job.Status {
		case Preparing:
			log.Debugf(ctx, "The job isn't published because it's preparing now: %v\n", job)
		case Ready:
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

func (m *Pipeline) HasNewTaskSince(ctx context.Context, t time.Time) (bool, error) {
	accessor := m.JobAccessor()
	q, err := accessor.Query()
	if err != nil {
		return false, err
	}
	q = q.Filter("CreatedAt >", t)
	c, err := q.Count(ctx)
	if err != nil {
		return false, err
	}
	return (c > 0), nil
}
