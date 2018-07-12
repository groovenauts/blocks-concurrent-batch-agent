package controller

import (
	"fmt"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func (c *PipelineBaseJobController) member(ctx context.Context, store *model.JobStore, idString string, RespondBadRequest func(error) error, RespondNotFound func(error) error, f func(*model.Job) error) error {
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return RespondBadRequest(fmt.Errorf("Invalid id: %q", idString))
	}
	m, err := store.ByID(ctx, id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return RespondNotFound(fmt.Errorf("Job not found: %q", id))
		} else {
			return err
		}
	}
	return f(m)
}
