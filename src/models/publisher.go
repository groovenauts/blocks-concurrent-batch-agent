package models

import (
	"golang.org/x/net/context"
	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, req *pubsub.PublishRequest) (string, error)
}

type PubsubPublisher struct {}

func (p *PubsubPublisher) Publish(ctx context.Context, topic string, req *pubsub.PublishRequest) (string, error) {
	// https://cloud.google.com/appengine/docs/standard/go/issue-requests
	client := urlfetch.Client(ctx)

	service, err := pubsub.New(client)
	if err != nil {
		log.Criticalf(ctx, "Failed to create pubsub.Service: %v\n", err)
		return "", err
	}

	call := service.Projects.Topics.Publish(topic, req)
	res, err := call.Do()
	if err != nil {
		log.Errorf(ctx, "Publish error: %v\n", err)
		return "", err
	}

	return res.MessageIds[0], nil
}

var GlobalPublisher Publisher = &PubsubPublisher{}
