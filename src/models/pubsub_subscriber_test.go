package models

import (
	"fmt"
	"testing"

	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/stretchr/testify/assert"

	"google.golang.org/appengine/aetest"
)

type DummyPuller struct {
	Error   error
	PullRes *pubsub.PullResponse
	AckRes  *pubsub.Empty
}

func (dp *DummyPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	if dp.Error != nil {
		return nil, dp.Error
	}
	res := dp.PullRes
	if res == nil {
		res = &pubsub.PullResponse{}
	}
	return dp.PullRes, nil
}

func (dp *DummyPuller) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	if dp.Error != nil {
		return nil, dp.Error
	}
	res := dp.AckRes
	if res == nil {
		res = &pubsub.Empty{}
	}
	return dp.AckRes, nil
}

func TestProcessProgressNotification(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	assert.NoError(t, err)
	defer done()

	dp := &DummyPuller{}

	ps := PubsubSubscriber{
		MessagePerPull: 1,
		puller:         dp,
	}

	subscription := "dummy-pipeline01-progress-subscription"

	recvMsg := &pubsub.ReceivedMessage{
		AckId: "dummy-ack-id",
	}

	returnNil := func(msg *pubsub.ReceivedMessage) error {
		return nil
	}

	returnError := func(msg *pubsub.ReceivedMessage) error {
		return fmt.Errorf("Dummy Error")
	}

	// Normal pattern
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.NoError(t, err)

	//  f returns an error
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnError)
	assert.Error(t, err)
	assert.Equal(t, "Dummy Error", err.Error())

	// Ack error and fail to get pipeline status
	dp.Error = fmt.Errorf("ack-error")
	err = ps.processProgressNotification(ctx, subscription, recvMsg, returnNil)
	assert.Error(t, err)
}
