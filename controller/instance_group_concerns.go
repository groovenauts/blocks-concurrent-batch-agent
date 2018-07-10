package controller

import (
	"fmt"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func (c *InstanceGroupController) member(ctx context.Context, store *model.InstanceGroupStore, idString string, RespondBadRequest func(error) error, RespondNotFound func(error) error, f func(*model.InstanceGroup) error) error {
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return RespondBadRequest(fmt.Errorf("Invalid id: %q", idString))
	}
	m, err := store.Get(ctx, id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return RespondNotFound(fmt.Errorf("InstanceGroup not found for %d", id))
		} else {
			return err
		}
	}
	return f(m)
}
