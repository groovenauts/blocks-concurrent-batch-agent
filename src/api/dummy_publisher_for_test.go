package api

import (
	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type DummyPublisher struct {
	ResultMessageId string
}

func (p *DummyPublisher) Publish(ctx context.Context, topic string, req *pubsub.PublishRequest) (string, error) {
	return p.ResultMessageId, nil
}
