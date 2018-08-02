package models

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	HibernationChecking:   "hibernation_checking",
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

	JobScaler struct {
		Enabled         bool `json:"enabled"`
		MaxInstanceSize int  `json:"max_instance_size"`
	}

	Pulling struct {
		MessagePerPull  int64 `json:"message_per_pull"`
		IntervalSeconds int64 `json:"interval_seconds"`
		JobsPerTask     int   `json:"jobs_per_task"`
	}

	Pipeline struct {
		ID                   string `json:"id"             datastore:"-"`
		key                  *datastore.Key
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
		DockerRunOptions     string         `json:"docker_run_options"`
		Status               Status         `json:"status"`
		Cancelled            bool           `json:"cancelled"`
		Dryrun               bool           `json:"dryrun"`
		DeploymentName       string         `json:"deployment_name"`
		TokenConsumption     int            `json:"token_consumption"`
		Dependency           Dependency     `json:"dependency,omitempty"`
		ClosePolicy          ClosePolicy    `json:"close_policy,omitempty"`
		HibernationDelay     int            `json:"hibernation_delay,omitempty"` // seconds
		HibernationStartedAt time.Time      `json:"hibernation_started_at,omitempty"`
		JobScaler            JobScaler      `json:"job_scaler,omitempty"`
		Pulling              Pulling        `json:"pulling"`
		PullingTaskSize      int            `json:"pulling_task_size"`
		InstanceSize         int            `json:"-"`
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

	if m.InstanceSize == 0 {
		m.InstanceSize = m.TargetSize
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
	m.key = res
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

	err := m.Validate()
	if err != nil {
		return err
	}

	key, err := datastore.DecodeKey(m.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to datastore.DecodeKey %q because of %v\n", m.ID, err)
		return err
	}
	_, err = datastore.Put(ctx, key, m)
	if err != nil {
		log.Errorf(ctx, "Failed to datastore.Put %v with key %v because of %v\n", m, key, err)
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
	log.Infof(ctx, "Now Pipeline status transition is going from %v to %v\n", m.Status, to)
	m.Status = to
	return m.Update(ctx)
}

func (m *Pipeline) StartBuilding(ctx context.Context) error {
	m.InstanceSize = m.TargetSize
	return m.StateTransition(ctx, []Status{Reserved, Building, Hibernating}, Building)
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
	return m.StateTransition(ctx, []Status{Opened, HibernationChecking}, HibernationChecking)
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
	m.InstanceSize = 0
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
	if err := m.LoadOrganization(ctx); err != nil {
		return err
	}
	if err := m.CancelLivingJobs(ctx); err != nil {
		return err
	}
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {

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
	return &JobAccessor{PipelineKey: m.key}
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
	accessor := m.JobAccessor()
	jobs, err := accessor.AllWith(ctx, func(q *datastore.Query) (*datastore.Query, error) {
		return q.Project("status"), nil
	})
	if err != nil {
		log.Errorf(ctx, "Failed to get jobs for %v because of %v\n", m.ID, err)
		return err
	}

	for _, j := range jobs {
		switch j.Status {
		case Preparing:
			log.Debugf(ctx, "The job isn't published because it's preparing now: %v\n", j)
		case Ready:
			jobId := j.ID
			err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
				job, err := accessor.Find(ctx, jobId)
				if err != nil {
					log.Warningf(ctx, "Failed to get job %v because of %v\n", jobId, err)
					return err
				}
				if job.Status != Ready {
					return nil
				}
				job.Pipeline = m
				_, err = job.Publish(ctx)
				if err != nil {
					log.Warningf(ctx, "Failed to publish job %v because of %v\n", job, err)
					return err
				}
				return nil
			}, nil)
			if err != nil {
				log.Errorf(ctx, "Failed to Publish job %v because of %v\n", jobId, err)
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

type ByPublishTime []*pubsub.ReceivedMessage

func (a ByPublishTime) Len() int      { return len(a) }
func (a ByPublishTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPublishTime) Less(i, j int) bool {
	return strings.Compare(a[i].Message.PublishTime, a[j].Message.PublishTime) < 0
}

type ErrorMessages []string

func (em ErrorMessages) Error() error {
	if len(em) > 0 {
		return fmt.Errorf(strings.Join(em, "\n"))
	}
	return nil
}

func (m *Pipeline) PullAndUpdateJobStatus(ctx context.Context) error {
	log.Infof(ctx, "PullAndUpdateJobStatus start\n")
	defer log.Infof(ctx, "PullAndUpdateJobStatus end\n")

	s := &PubsubSubscriber{MessagePerPull: Int64WithDefault(m.Pulling.MessagePerPull, 100)}
	err := s.setup(ctx)
	if err != nil {
		return err
	}

	// log.Debugf(ctx, "PullAndUpdateJobStatus #1\n")

	messagesForJob := map[string][]*pubsub.ReceivedMessage{}

	accessor := m.JobAccessor()
	subscription := m.ProgressSubscriptionFqn()
	err = s.subscribe(ctx, subscription, func(recvMsg *pubsub.ReceivedMessage) error {
		attrs := recvMsg.Message.Attributes
		jobId := attrs[JobIdKey]
		if messagesForJob[jobId] == nil {
			messages := []*pubsub.ReceivedMessage{}
			messagesForJob[jobId] = messages
		}
		messagesForJob[jobId] = append(messagesForJob[jobId], recvMsg)
		return nil
	})
	if err != nil {
		return err
	}

	// log.Debugf(ctx, "PullAndUpdateJobStatus #2\n")

	for _, recvMsgs := range messagesForJob {
		sort.Sort(ByPublishTime(recvMsgs))
	}

	// log.Debugf(ctx, "PullAndUpdateJobStatus #3\n")

	errors := ErrorMessages{}
	for jobId, recvMsgs := range messagesForJob {
		err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			// log.Debugf(ctx, "PullAndUpdateJobStatus #4.1\n")
			job, err := accessor.Find(ctx, jobId)
			if err != nil {
				return err
			}
			// log.Debugf(ctx, "PullAndUpdateJobStatus #4.2\n")

			if err := m.OverwriteJobByMessages(ctx, job, recvMsgs); err != nil {
				return err
			}
			// log.Debugf(ctx, "PullAndUpdateJobStatus #4.3\n")

			if err := job.Update(ctx); err != nil {
				return err
			}
			// log.Debugf(ctx, "PullAndUpdateJobStatus #4.4\n")
			return nil
		}, &datastore.TransactionOptions{Attempts: 10, XG: true})
		if err != nil {
			if err == datastore.ErrConcurrentTransaction {
				return err
			}
			errors = append(errors, err.Error())
		}
		// log.Debugf(ctx, "PullAndUpdateJobStatus #4.5\n")
		for _, recvMsg := range recvMsgs {
			err := s.sendAck(ctx, subscription, recvMsg)
			if err != nil {
				log.Warningf(ctx, "Failed to send ACK(%v) to %q because of %v. It will be delivered later again.\n", recvMsg.AckId, subscription, err)
				errors = append(errors, err.Error())
			}
		}
	}

	return errors.Error()
}

func (m *Pipeline) OverwriteJobByMessages(ctx context.Context, job *Job, recvMsgs []*pubsub.ReceivedMessage) error {
	errors := ErrorMessages{}
	for _, recvMsg := range recvMsgs {
		err := m.OverwriteJob(ctx, job, recvMsg)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	return errors.Error()
}

func (m *Pipeline) OverwriteJob(ctx context.Context, job *Job, recvMsg *pubsub.ReceivedMessage) error {
	if len(recvMsg.Message.Data) > 0 {
		b, err := base64.StdEncoding.DecodeString(recvMsg.Message.Data)
		if err != nil {
			return err
		}
		job.Output += "\n\n" + string(b)
	}

	if job.Status == Success {
		return nil
	}

	attrs := recvMsg.Message.Attributes
	completed, err := strconv.ParseBool(attrs["completed"])
	if err != nil {
		return err
	}
	if completed {
		job.Status = Success
	}

	step, err := ParseJobStep(attrs["step"])
	if err != nil {
		return err
	}
	stepStatus, err := ParseJobStepStatus(attrs["step_status"])
	if err != nil {
		return err
	}
	job.Hostname = m.stringFromMapWithDefault(attrs, "host", "unknown")
	job.Zone = m.stringFromMapWithDefault(attrs, "zone", "unknown")
	job.StartTime = m.stringFromMapWithDefault(attrs, "job.start-time", "")
	job.FinishTime = m.stringFromMapWithDefault(attrs, "job.finish-time", "")

	log.Debugf(ctx, "PullAndUpdateJobStatus len(recvMsg.Message.Data): %v\n", len(recvMsg.Message.Data))

	job.ApplyStatusIfGreaterThanBefore(ctx, completed, step, stepStatus)
	return nil
}

func (m *Pipeline) stringFromMapWithDefault(src map[string]string, key, defaultValue string) string {
	r, ok := src[key]
	if !ok {
		return defaultValue
	}
	return r
}

func (m *Pipeline) JobCountBy(ctx context.Context, st JobStatus) (int, error) {
	accessor := m.JobAccessor()
	q := accessor.Query()
	r, err := q.Filter("status = ", int(st)).Count(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to get Count by query %v because of %v\n", q, err)
		return 0, err
	}
	return r, nil
}

func (m *Pipeline) JobCount(ctx context.Context, statuses ...JobStatus) (int, error) {
	r := 0
	for _, st := range statuses {
		c, err := m.JobCountBy(ctx, st)
		if err != nil {
			return 0, err
		}
		r += c
	}
	return r, nil
}

func (m *Pipeline) CalcAndUpdatePullingTaskSize(ctx context.Context, jobCount int, f func(int) error) error {
	jobsPerTask := m.Pulling.JobsPerTask
	if jobsPerTask < 1 {
		jobsPerTask = 50
	}
	tasks := (jobCount / jobsPerTask) + 1
	newTasks := tasks - m.PullingTaskSize
	if newTasks > 0 {
		m.PullingTaskSize = tasks
		if err := m.Update(ctx); err != nil {
			return err
		}
		return f(newTasks)
	}
	return nil
}

func (m *Pipeline) DecreasePullingTaskSize(ctx context.Context, diff int, f func() error) error {
	ret := f()
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := m.Reload(ctx); err != nil {
			log.Warningf(ctx, "Failed to reload on Pipeline.DecreasePullingTaskSize for %v because of %v\n", m.ID, err)
			return err
		}
		m.PullingTaskSize = -diff
		if err := m.Update(ctx); err != nil {
			log.Warningf(ctx, "Failed to update on Pipeline.DecreasePullingTaskSize for %v because of %v\n", m.ID, err)
			return err
		}
		return nil
	}, &datastore.TransactionOptions{XG: true})
	if err != nil {
		log.Errorf(ctx, "Failed to update on Pipeline.DecreasePullingTaskSize for %v because of %v\n", m.ID, err)
	}
	return ret
}

func (m *Pipeline) HasNewTaskSince(ctx context.Context, t time.Time) (bool, error) {
	accessor := m.JobAccessor()
	q := accessor.Query()
	q = q.Filter("CreatedAt >", t)
	c, err := q.Count(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to get Count with query %v because of %v\n", q, err)
		return false, err
	}
	return (c > 0), nil
}

func (m *Pipeline) CanScale() bool {
	return m.JobScaler.Enabled && (m.InstanceSize < m.JobScaler.MaxInstanceSize)
}

func (m *Pipeline) LogInstanceSizeWithError(ctx context.Context, endTime string, size int) error {
	var t time.Time
	if endTime != "" {
		var err error
		t, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			t = time.Now()
			log.Warningf(ctx, "ERROR failed to time.Parse %v with RFC3339, so use time.Now: %v\n", endTime, t)
		}
	}

	sizeLog := &PipelineInstanceSizeLog{
		pipeline:  m,
		Size:      size,
		CreatedAt: t,
	}
	return sizeLog.Create(ctx)
}

func (m *Pipeline) LogInstanceSize(ctx context.Context, endTime string, size int) {
	if err := m.LogInstanceSizeWithError(ctx, endTime, size); err != nil {
		log.Warningf(ctx, "ERROR failed to insert to PipelineInstanceSizeLogs Pipeline.ID: %v, Size: %v, CreatedAt: %v because of %v\n", m.ID, size, endTime, err)
	}
}
