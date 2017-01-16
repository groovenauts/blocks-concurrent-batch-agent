package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Resizer struct {
}

func (b *Resizer) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Resizing pipeline %v\n", pl)
	return nil
}
