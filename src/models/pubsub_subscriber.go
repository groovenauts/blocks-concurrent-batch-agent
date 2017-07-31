package models

import (
	pubsub "google.golang.org/api/pubsub/v1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

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

func (ps *PubsubSubscriber) subscribe(ctx context.Context, subscription *Subscription, f func(msg *pubsub.ReceivedMessage) error) error {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: true,
		MaxMessages:       ps.MessagePerPull,
	}
	log.Debugf(ctx, "%v Pulling subscription\n", subscription.Name)
	res, err := ps.puller.Pull(subscription.Name, pullRequest)
	if err != nil {
		log.Errorf(ctx, "%v Failed to pull: [%T] %v\n", subscription.Name, err, err)
		return err
	}
	log.Debugf(ctx, "%v Pulled successfully\n", subscription.Name)
	for _, receivedMessage := range res.ReceivedMessages {
		err := ps.processProgressNotification(ctx, subscription, receivedMessage, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PubsubSubscriber) processProgressNotification(ctx context.Context, subscription *Subscription, receivedMessage *pubsub.ReceivedMessage, f func(msg *pubsub.ReceivedMessage) error) error {
	err := f(receivedMessage)
	if err != nil {
		log.Errorf(ctx, "the received request process returns error: [%T] %v", err, err)
		return err
	}
	_, err = ps.puller.Acknowledge(subscription.Name, receivedMessage.AckId)
	if err != nil {
		log.Errorf(ctx, "Failed to acknowledge for message: %v cause of [%T] %v", receivedMessage, err, err)
		return err
	}
	return nil
}
