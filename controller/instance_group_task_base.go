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
	RemoteOpeFunc func(context.Context, *model.CloudAsyncOperation) (model.RemoteOperationWrapper, error)
	WatchTaskPathFunc func(*model.CloudAsyncOperation) string
	RespondOK func(*app.CloudAsyncOperation) error
	RespondAccepted func(*app.CloudAsyncOperation) error
	RespondNoContent func(*app.CloudAsyncOperation) error
	RespondCreated func(*app.CloudAsyncOperation) error
}

// Start
func (t *InstanceGroupTaskBase) Start(appCtx context.Context, resourceIdString string) error {
	resourceId, err := strconv.ParseInt(resourceIdString, 10, 64)
	if err != nil {
		log.Errorf(appCtx, "Invalid resource ID: %q\n", resourceIdString)
		return t.RespondNoContent(nil)
	}

	store := &model.InstanceGroupStore{}
	m, err := store.Get(appCtx, resourceId)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "InstanceGroup not found for %q\n", resourceId)
			return t.RespondNoContent(nil)
		} else {
			return err
		}
	}

	if m.Status != t.MainStatus {
		if t.IsSkipped(m.Status) {
			log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return t.RespondOK(nil)
		} else {
			log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return t.RespondNoContent(nil)
		}
	}

	processor, err := t.ProcessorFactory(appCtx)
	ope, err := processor.Process(appCtx, m)
	if err != nil {
		return err
	}
	opeStore := &model.CloudAsyncOperationStore{}
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
		return t.RespondCreated(CloudAsyncOperationModelToMediaType(ope))
	}, nil)
}

// Watch
func (t *InstanceGroupTaskBase) Watch(appCtx context.Context, opeIdString string) error {
	opeId, err := strconv.ParseInt(opeIdString, 10, 64)
	if err != nil {
		log.Errorf(appCtx, "Invalid operation ID: %q\n", opeIdString)
		return t.RespondNoContent(nil)
	}

	opeStore := &model.CloudAsyncOperationStore{}
	ope, err := opeStore.Get(appCtx, opeId)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "CloudAsyncOperation not found for %q\n", opeId)
			return t.RespondNoContent(nil)
		} else {
			return err
		}
	}
	store := &model.InstanceGroupStore{}
	m, err := store.Get(appCtx, ope.OwnerID)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(appCtx, "InstanceGroup not found for %q\n", ope.OwnerID)
			return t.RespondNoContent(nil)
		} else {
			return err
		}
	}

	if m.Status != t.MainStatus {
		if t.IsSkipped(m.Status) {
			log.Infof(appCtx, "SKIPPING because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return t.RespondOK(nil)
		} else {
			log.Warningf(appCtx, "Invalid request because InstanceGroup %s is already %v\n", m.Id, m.Status)
			return t.RespondNoContent(nil)
		}
	}

	remoteOpe, err := t.RemoteOpeFunc(appCtx, ope)
	if err != nil {
		return err
	}

	if ope.Status != remoteOpe.Status() {
		ope.AppendLog(fmt.Sprintf("InstanceGroup %q Status changed from %q to %q", m.Id, ope.Status, remoteOpe.Status()))
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
		return t.RespondCreated(CloudAsyncOperationModelToMediaType(ope))
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
		return f(CloudAsyncOperationModelToMediaType(ope))
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
