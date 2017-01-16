package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Closer struct {
}

func (b *Closer) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Closing pipeline %v\n", pl)
	return nil
}
