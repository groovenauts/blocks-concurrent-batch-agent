package controller

import (
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
	SkipStatuses []model.InstanceGroupStatus
	ProcessorFactory func(ctx context.Context) (model.InstanceGroupProcessor, error)
	WatchTaskPathFunc func(*model.CloudAsyncOperation) string
	RespondOk func(*app.CloudAsyncOperation) error
	RespondNoContent func(*app.CloudAsyncOperation) error
	RespondCreated func(*app.CloudAsyncOperation) error
}

func (t *InstanceGroupTaskBase) Start(appCtx context.Context, resourceId string) error {
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
			return t.RespondOk(nil)
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

func (t *InstanceGroupTaskBase) IsSkipped(status model.InstanceGroupStatus) bool {
	for _, st := range t.SkipStatuses {
		if status == st {
			return true
		}
	}
	return false
}
