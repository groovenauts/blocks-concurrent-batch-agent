package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func (c *JobController) member(ctx context.Context, store *model.JobStore, id string, RespondNotFound func(error) error, f func(*model.Job) error) error {
	m, err := store.Get(ctx, id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return RespondNotFound(fmt.Errorf("Job not found: %q", id))
		} else {
			return err
		}
	}
	return f(m)
}
