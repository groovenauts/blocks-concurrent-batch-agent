package models

import (
	"fmt"

	"golang.org/x/net/context"
	pubsub "google.golang.org/api/pubsub/v1"
)

type PublishInvocation struct {
	Topic string
	Req   *pubsub.PublishRequest
}

type DummyPublisher struct {
	Invocations []*PublishInvocation
}

func (p *DummyPublisher) Publish(ctx context.Context, topic string, req *pubsub.PublishRequest) (string, error) {
	p.Invocations = append(p.Invocations, &PublishInvocation{Topic: topic, Req: req})
	return fmt.Sprintf("DummyMsgId-%v", len(p.Invocations)), nil
}
