package controller

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/net/context"

	// "google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	// "github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

type InstanceGroupTaskBase struct {
	MainStatus        model.InstanceGroupStatus
	NextStatus        model.InstanceGroupStatus
	ErrorStatus       model.InstanceGroupStatus
	SkipStatuses      []model.InstanceGroupStatus
	ProcessorFactory  func(ctx context.Context) (model.InstanceGroupProcessor, error)
	RemoteOpeFunc     func(context.Context, *model.InstanceGroupOperation) (model.RemoteOperationWrapper, error)
	WatchTaskPathFunc func(*model.InstanceGroupOperation) string
	RespondOK         func(*app.CloudAsyncOperation) error
	RespondAccepted   func(*app.CloudAsyncOperation) error
	RespondNoContent  func(*app.CloudAsyncOperation) error
	RespondCreated    func(*app.CloudAsyncOperation) error
}

func (t *InstanceGroupTaskBase) WithInstanceGroupStore(ctx context.Context, orgIdString, name string, f func(*model.InstanceGroupStore, *model.InstanceGroupOperationStore) error) error {
	orgId, err := strconv.ParseInt(orgIdString, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid Organization ID: %v", orgIdString)
	}

	goon := model.GoonFromContext(ctx)

	orgKey, err := goon.KeyError(&model.Organization{ID: orgId})
	if err != nil {
		return fmt.Errorf("Can't get Organization key from %v", orgId)
	}

	igStore := &model.InstanceGroupStore{ParentKey: orgKey}

	igKey, err := goon.KeyError(&model.InstanceGroup{ParentKey: orgKey, Name: name})
	if err != nil {
		return fmt.Errorf("Can't get InstanceGroup key from %d / %v", orgId, name)
	}

	opeStore := &model.InstanceGroupOperationStore{ParentKey: igKey}

	return f(igStore, opeStore)
}

// Start
func (t *InstanceGroupTaskBase) Start(appCtx context.Context, orgId, name string) error {
	return t.WithInstanceGroupStore(appCtx, orgId, name, func(igStore *model.InstanceGroupStore, opeStore *model.InstanceGroupOperationStore) error {
		m, err := igStore.ByID(appCtx, name)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				log.Errorf(appCtx, "InstanceGroup not found for %q\n", name)
				return t.RespondNoContent(nil)
			} else {
				return err
			}
		}

		if m.Status != t.MainStatus {
			if t.IsSkipped(m.Status) {
				log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Name, m.Status)
				return t.RespondOK(nil)
			} else {
				log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Name, m.Status)
				return t.RespondNoContent(nil)
			}
		}

		processor, err := t.ProcessorFactory(appCtx)
		ope, err := processor.Process(appCtx, m)
		if err != nil {
			return err
		}

		_, err = opeStore.Put(appCtx, ope)
		if err != nil {
			return err
		}

		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			m.Status = t.NextStatus
			_, err = igStore.Put(c, m)
			if err != nil {
				return err
			}
			path := t.WatchTaskPathFunc(ope)
			if err := PutTask(appCtx, path, 0); err != nil {
				return err
			}
			return t.RespondCreated(InstanceGroupOperationModelToMediaType(ope))
		}, nil)
	})
}

// Watch
func (t *InstanceGroupTaskBase) Watch(appCtx context.Context, orgId, name, opeIdString string) error {
	return t.WithInstanceGroupStore(appCtx, orgId, name, func(igStore *model.InstanceGroupStore, opeStore *model.InstanceGroupOperationStore) error {
		opeId, err := strconv.ParseInt(opeIdString, 10, 64)
		if err != nil {
			log.Errorf(appCtx, "Invalid operation ID: %q\n", opeIdString)
			return t.RespondNoContent(nil)
		}

		ope, err := opeStore.ByID(appCtx, opeId)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				log.Errorf(appCtx, "InstanceGroupOperation not found for %q\n", opeId)
				return t.RespondNoContent(nil)
			} else {
				return err
			}
		}
		m, err := igStore.ByKey(appCtx, ope.ParentKey)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				log.Errorf(appCtx, "InstanceGroup not found for %q\n", ope.Id)
				return t.RespondNoContent(nil)
			} else {
				return err
			}
		}

		if m.Status != t.MainStatus {
			if t.IsSkipped(m.Status) {
				log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Name, m.Status)
				return t.RespondOK(nil)
			} else {
				log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Name, m.Status)
				return t.RespondNoContent(nil)
			}
		}

		remoteOpe, err := t.RemoteOpeFunc(appCtx, ope)
		if err != nil {
			return err
		}

		if ope.Status != remoteOpe.Status() {
			ope.AppendLog(fmt.Sprintf("InstanceGroup %q Status changed from %q to %q", m.Name, ope.Status, remoteOpe.Status()))
		}

		// PENDING, RUNNING, or DONE
		switch remoteOpe.Status() {
		case "DONE": // through
		default:
			if ope.Status != remoteOpe.Status() {
				ope.Status = remoteOpe.Status()
				_, err := opeStore.Update(appCtx, ope)
				if err != nil {
					return err
				}
			}
			path := t.WatchTaskPathFunc(ope)
			if err := PutTask(appCtx, path, 1*time.Minute); err != nil {
				return err
			}
			return t.RespondCreated(InstanceGroupOperationModelToMediaType(ope))
		}

		errors := remoteOpe.Errors()
		var f func(r *app.CloudAsyncOperation) error
		if errors != nil {
			ope.Errors = *errors
			ope.AppendLog(fmt.Sprintf("Error by %v", remoteOpe.GetOriginal()))
			m.Status = t.ErrorStatus
			f = t.RespondNoContent
		} else {
			ope.AppendLog("Success")
			m.Status = t.NextStatus
			f = t.RespondAccepted
		}

		_, err = opeStore.Update(appCtx, ope)
		if err != nil {
			return err
		}

		return datastore.RunInTransaction(appCtx, func(c context.Context) error {
			_, err = igStore.Update(c, m)
			if err != nil {
				return err
			}
			// TODO Add calling PipelineBase callback
			return f(InstanceGroupOperationModelToMediaType(ope))
		}, nil)
	})
}

func (t *InstanceGroupTaskBase) IsSkipped(status model.InstanceGroupStatus) bool {
	return t.IncludedStatus(status, t.SkipStatuses)
}

func (t *InstanceGroupTaskBase) IncludedStatus(status model.InstanceGroupStatus, statuses []model.InstanceGroupStatus) bool {
	for _, st := range statuses {
		if status == st {
			return true
		}
	}
	return false
}
