package pipeline

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	// "google.golang.org/appengine/log"
)

type Closer struct {
}

func (b *Closer) Process(ctx context.Context, pl *Pipeline) error {
	pl.Props.Status = closed
	key, err := datastore.DecodeKey(pl.ID)
	if err != nil {
		return err
	}
	_, err = datastore.Put(ctx, key, &pl.Props)
	if err != nil {
		return err
	}
	return nil
}
