package models

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"gopkg.in/go-playground/validator.v9"
)

// Status constants
type Status int

const (
	Initialized Status = iota
	Broken
	Pending
	Reserved
	Building
	Deploying
	Opened
	Closing
	Closing_error
	Closed
)

var StatusStrings = map[Status]string{
	Initialized:   "initialized",
	Broken:        "broken",
	Pending:       "pending",
	Reserved:      "reserved",
	Building:      "building",
	Deploying:     "deploying",
	Opened:        "opened",
	Closing:       "closing",
	Closing_error: "closing_error",
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
		CreatedAt              time.Time         `json:"created_at"`
		UpdatedAt              time.Time         `json:"updated_at"`
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

func (m *Pipeline) ReserveOrWait(ctx context.Context) error {
	return m.CreateWith(ctx, func(ctx context.Context) error{
		err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
			if err != nil {
				return err
			}

			pending, err := org.PipelineAccessor().PendingQuery()
			if err != nil {
				return err
			}

			cnt, err := pending.Count(ctx)
			if err != nil {
				return err
			}

			if cnt > 0 {
				log.Warningf(ctx, "Insufficient tokens; %v has already %v pending pipelines", org.Name, cnt)
				m.Status = Pending
			} else {
				newAmount := org.TokenAmount - m.TokenConsumption
				if newAmount < 0 {
					log.Warningf(ctx, "Insufficient tokens; %v has only %v tokens but %v required %v tokens", org.Name, org.TokenAmount, m.Name, m.TokenConsumption)
					m.Status = Pending
				} else {
					m.Status = Reserved
					org.TokenAmount = newAmount
					err = org.Update(ctx)
					if err != nil {
						return err
					}
				}
			}

			return m.PutWithNewKey(ctx)
		}, nil)

		if err != nil {
			log.Errorf(ctx, "Transaction failed: %v\n", err)
			return err
		}

		return nil
	})
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
	m.UpdatedAt = time.Now()

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

func (m *Pipeline) RefreshHandler(ctx context.Context) func(*[]DeploymentError)error {
	return m.RefreshHandlerWith(ctx, nil)
}

func (m *Pipeline) RefreshHandlerWith(ctx context.Context, pipelineProcesser func(*Pipeline) error) func(*[]DeploymentError)error {
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

func (m *Pipeline) FailDeploying(ctx context.Context, errors *[]DeploymentError) error {
	m.DeployingErrors = *errors
	m.Status = Broken
	return m.Update(ctx)
}

func (m *Pipeline) CompleteDeploying(ctx context.Context) error {
	m.Status = Opened
	return m.Update(ctx)
}

func (m *Pipeline) FailClosing(ctx context.Context, errors *[]DeploymentError) error {
	m.ClosingErrors = *errors
	m.Status = Closing_error
	return m.Update(ctx)
}

func (m *Pipeline) CompleteClosing(ctx context.Context, pipelineProcesser func(*Pipeline) error) error {
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		org, err := GlobalOrganizationAccessor.Find(ctx, m.Organization.ID)
		if err != nil {
			return err
		}
		newTokenAmount := org.TokenAmount + m.TokenConsumption

		pendings, err := org.PipelineAccessor().GetPendings(ctx)
		if err != nil {
			return err
		}

		for _, pending := range pendings {
			if newTokenAmount < pending.TokenConsumption {
				break
			}
			newTokenAmount = newTokenAmount - pending.TokenConsumption
			pending.Status = Reserved
			err := pending.Update(ctx)
			if err != nil {
				return err
			}
			if pipelineProcesser != nil {
				err := pipelineProcesser(pending)
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

		m.Status = Closed
		return m.Update(ctx)
	}, nil)
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

func (m *Pipeline) JobAccessor() *PipelineJobAccessor {
	return &PipelineJobAccessor{Parent: m}
}

func (m *Pipeline) Reload(ctx context.Context) error {
	err := GlobalPipelineAccessor.LoadByID(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Pipeline) WaitUntil(ctx context.Context, st Status, interval, timeout time.Duration) error {
	t0 := time.Now()
	deadline := t0.Add(timeout)
	for deadline.After(time.Now()) {
		m.Reload(ctx)
		if m.Status == st {
			return nil
		}
		time.Sleep(interval)
	}
	return ErrTimeout
}

func (m *Pipeline) JobTopicName() string {
	return fmt.Sprintf("%s-job-topic", m.Name)
}

func (m *Pipeline) JobTopicFqn() string {
	return fmt.Sprintf("projects/%s/topics/%s", m.ProjectID, m.JobTopicName())
}

func (m *Pipeline) IDHex() string {
	return hex.EncodeToString([]byte(m.ID))
}
