package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func (c *InstanceGroupController) member(ctx context.Context, store *model.InstanceGroupStore, name string, RespondBadRequest func(error) error, RespondNotFound func(error) error, f func(*model.InstanceGroup) error) error {
	m, err := store.ByID(ctx, name)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return RespondNotFound(fmt.Errorf("InstanceGroup not found for %d", name))
		} else {
			return err
		}
	}
	return f(m)
}
