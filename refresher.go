package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Refresher struct {
}

func (b *Refresher) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Refreshing pipeline %v\n", pl)
	return nil
}
