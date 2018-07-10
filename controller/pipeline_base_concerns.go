package controller

import (
	"fmt"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func (c *PipelineBaseController) member(ctx context.Context, store *model.PipelineBaseStore, idString string, RespondBadRequest func(error) error, RespondNotFound func(error) error, f func(*model.PipelineBase) error) error {
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return RespondBadRequest(fmt.Errorf("Invalid id: %q", idString))
	}
	m, err := store.Get(ctx, id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return RespondNotFound(fmt.Errorf("PipelineBase not found: %q", id))
		} else {
			return err
		}
	}
	return f(m)
}
