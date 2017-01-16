package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Updater struct {
}

func (b *Updater) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Updating pipeline %v\n", pl)
	return nil
}
