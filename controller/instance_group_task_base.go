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
	MainStatus model.InstanceGroupStatus
	NextStatus model.InstanceGroupStatus
	ErrorStatus model.InstanceGroupStatus
	SkipStatuses []model.InstanceGroupStatus
	ProcessorFactory func(ctx context.Context) (model.InstanceGroupProcessor, error)
	RemoteOpeFunc func(context.Context, *model.InstanceGroupOperation) (model.RemoteOperationWrapper, error)
	WatchTaskPathFunc func(*model.InstanceGroupOperation) string
	RespondOK func(*app.CloudAsyncOperation) error
	RespondAccepted func(*app.CloudAsyncOperation) error
	RespondNoContent func(*app.CloudAsyncOperation) error
	RespondCreated func(*app.CloudAsyncOperation) error
}

// Start
func (t *InstanceGroupTaskBase) Start(appCtx context.Context, name string) error {
	store := &model.InstanceGroupStore{}
	m, err := store.ByID(appCtx, name)
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
	opeStore := &model.InstanceGroupOperationStore{}
	_, err = opeStore.Put(appCtx, ope)
	if err != nil {
		return err
	}

	return datastore.RunInTransaction(appCtx, func(c context.Context) error {
		m.Status = t.NextStatus
		_, err = store.Put(c, m)
		if err != nil {
			return err
		}
		path := t.WatchTaskPathFunc(ope)
		if err := PutTask(appCtx, path, 0); err != nil {
			return err
		}
		return t.RespondCreated(InstanceGroupOperationModelToMediaType(ope))
	}, nil)
}

// Watch
func (t *InstanceGroupTaskBase) Watch(appCtx context.Context, name, opeIdString string) error {
	opeId, err := strconv.ParseInt(opeIdString, 10, 64)
	if err != nil {
		log.Errorf(appCtx, "Invalid operation ID: %q\n", opeIdString)
		return t.RespondNoContent(nil)
	}

	opeStore := &model.InstanceGroupOperationStore{}
	ope, err := opeStore.ByID(appCtx, opeId)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "InstanceGroupOperation not found for %q\n", opeId)
			return t.RespondNoContent(nil)
		} else {
			return err
		}
	}
	store := &model.InstanceGroupStore{}
	m, err := store.ByKey(appCtx, ope.ParentKey)
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
		if err := PutTask(appCtx, path, 1 * time.Minute); err != nil {
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
		_, err = store.Update(c, m)
		if err != nil {
			return err
		}
		// TODO Add calling PipelineBase callback
		return f(InstanceGroupOperationModelToMediaType(ope))
	}, nil)
}

func (t *InstanceGroupTaskBase) IsSkipped(status model.InstanceGroupStatus) bool {
	for _, st := range t.SkipStatuses {
		if status == st {
			return true
		}
	}
	return false
}
