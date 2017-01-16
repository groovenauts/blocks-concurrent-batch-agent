package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type Builder struct {
}

func (b *Builder) Process(ctx context.Context, pl *Pipeline) error {
	log.Debugf(ctx, "Building pipeline %v\n", pl)
	return nil
}
