package models

import (
	"context"

	pubsub "google.golang.org/api/pubsub/v1"

	"golang.org/x/oauth2/google"

	"google.golang.org/api/googleapi"
	"google.golang.org/appengine/log"
)

type (
	Puller interface {
		Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error)
		Acknowledge(subscription, ackId string) (*pubsub.Empty, error)
	}

	pubsubPuller struct {
		subscriptionsService *pubsub.ProjectsSubscriptionsService
	}

	PubsubSubscriber struct {
		MessagePerPull int64
		puller         Puller
	}
)

func (pp *pubsubPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	return pp.subscriptionsService.Pull(subscription, pullrequest).Do()
}

func (pp *pubsubPuller) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	ackRequest := &pubsub.AcknowledgeRequest{
		AckIds: []string{ackId},
	}
	return pp.subscriptionsService.Acknowledge(subscription, ackRequest).Do()
}

func (ps *PubsubSubscriber) setup(ctx context.Context) error {
	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)
	if err != nil {
		log.Errorf(ctx, "Failed to create DefaultClient\n")
		return err
	}

	// Creates a pubsubClient
	service, err := pubsub.New(client)
	if err != nil {
		log.Errorf(ctx, "Failed to create pubsub.Service with %v: %v\n", client, err)
		return err
	}

	ps.puller = &pubsubPuller{service.Projects.Subscriptions}
	return nil
}

func (ps *PubsubSubscriber) subscribeAndAck(ctx context.Context, subscription string, f func(*pubsub.ReceivedMessage) error) error {
	return ps.subscribe(ctx, subscription, func(receivedMessage *pubsub.ReceivedMessage) error {
		return ps.processProgressNotification(ctx, subscription, receivedMessage, f)
	})
}

func (ps *PubsubSubscriber) subscribe(ctx context.Context, subscription string, f func(*pubsub.ReceivedMessage) error) error {
	log.Infof(ctx, "PubsubSubscriber.subscribe start\n")
	defer log.Infof(ctx, "PubsubSubscriber.subscribe end\n")

	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: true,
		MaxMessages:       ps.MessagePerPull,
	}
	log.Debugf(ctx, "%v Pulling subscription\n", subscription)
	res, err := ps.puller.Pull(subscription, pullRequest)
	if err != nil {
		switch err.(type) {
		case *googleapi.Error:
			apiError := err.(*googleapi.Error)
			if apiError.Code == 404 {
				return &SubscriprionNotFound{Subscription: subscription}
			}
		}
		log.Errorf(ctx, "%v Failed to pull: [%T] %v\n", subscription, err, err)
		return err
	}
	log.Debugf(ctx, "%v Pulled %v messages successfully.", subscription, len(res.ReceivedMessages))
	for i, recv := range res.ReceivedMessages {
		m := recv.Message
		log.Debugf(ctx, "Pulled Message #%v AckId: %v, MessageId: %v, PublishTime: %v, Attributes: %v, Data: %v\n", i, recv.AckId, m.MessageId, m.PublishTime, m.Attributes, m.Data)
	}

	for _, recv := range res.ReceivedMessages {
		// m := recv.Message
		// log.Debugf(ctx, "Pulled Message #%v AckId: %v, MessageId: %v, PublishTime: %v\n", i, recv.AckId, m.MessageId, m.PublishTime)
		err := f(recv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PubsubSubscriber) processProgressNotification(ctx context.Context, subscription string, receivedMessage *pubsub.ReceivedMessage, f func(*pubsub.ReceivedMessage) error) error {
	err := f(receivedMessage)
	if err != nil {
		log.Errorf(ctx, "the received request process returns error: [%T] %v", err, err)
		return err
	}
	return ps.sendAck(ctx, subscription, receivedMessage)
}

func (ps *PubsubSubscriber) sendAck(ctx context.Context, subscription string, receivedMessage *pubsub.ReceivedMessage) error {
	_, err := ps.puller.Acknowledge(subscription, receivedMessage.AckId)
	if err != nil {
		log.Errorf(ctx, "Failed to acknowledge for message: %v cause of [%T] %v", receivedMessage, err, err)
		return err
	}
	return nil
}
