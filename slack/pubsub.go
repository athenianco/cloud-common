package slack

import (
	"context"

	"github.com/athenianco/cloud-common/pubsub"
)

func NewPubSub(topic string) (Client, error) {
	p, err := pubsub.NewPublisher(topic)
	if err != nil {
		return nil, err
	}
	return &pubsubClient{p: p}, nil
}

type pubsubClient struct {
	p pubsub.Publisher
}

func (c *pubsubClient) SendMessage(ctx context.Context, m Message) error {
	return pubsub.PublishJSON(ctx, c.p, m)
}
