package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func (c *InstanceGroupController) member(ctx context.Context, store *model.InstanceGroupStore, id string, RespondNotFound func(error) error, f func(*model.InstanceGroup) error) error {
	m, err := store.Get(ctx, id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return RespondNotFound(fmt.Errorf("InstanceGroup not found: %q", id))
		} else {
			return err
		}
	}
	return f(m)
}
