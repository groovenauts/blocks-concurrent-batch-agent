package model

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/appengine/log"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, req *pubsub.PublishRequest) (string, error)
}

type PubsubPublisher struct{}

func (p *PubsubPublisher) Publish(ctx context.Context, topic string, req *pubsub.PublishRequest) (string, error) {
	// https://developers.google.com/identity/protocols/application-default-credentials#callinggo
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)
	if err != nil {
		log.Criticalf(ctx, "Failed to get google.DefaultClient for pubsub scope because of %v\n", err)
		return "", err
	}

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
