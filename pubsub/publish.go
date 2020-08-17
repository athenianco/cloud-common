package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	gpubsub "cloud.google.com/go/pubsub"
	"github.com/athenianco/cloud-common/gcp"
	"github.com/athenianco/cloud-common/report"
)

var checkTopics = os.Getenv("ATHENIAN_CHECK_TOPICS") == "true"

var _ Publisher = (*gcpPublisher)(nil)

// gcpPublisher is Google Pub/Sub publisher.
type gcpPublisher struct {
	topic *gpubsub.Topic
}

// NewPublisherFromEnv is similar to NewPublisher, but takes
// the topic name from the given environment variable(s).
func NewPublisherFromEnv(envs ...string) (Publisher, error) {
	for _, env := range envs {
		if topic := os.Getenv(env); topic != "" {
			return NewPublisher(topic)
		}
	}
	if len(envs) == 1 {
		return nil, errors.New(envs[0] + " must be specified")
	}
	return nil, fmt.Errorf("one of the %s must be specified", strings.Join(envs, ", "))
}

// NewPublisher creates a new instance of Pub/Sub publisher.
// It also creates the Pub/Sub topic if it does not exist.
func NewPublisher(topicID string) (Publisher, error) {
	ctx := context.Background()

	client, err := gpubsub.NewClient(ctx, gcp.ProjectID())
	if err != nil {
		report.Error(ctx, err)
		return nil, err
	}

	topic := client.Topic(topicID)
	if checkTopics {
		exists, err := topic.Exists(ctx)
		if err != nil {
			report.Error(ctx, err)
			return nil, err
		} else if !exists {
			err = fmt.Errorf("topic doesn't exist: %q", topicID)
			report.Error(ctx, err)
			return nil, err
		}
	}
	return &gcpPublisher{topic: topic}, nil
}

// Publish messages to the Pub/Sub topic synchronously.
func (p *gcpPublisher) Publish(ctx context.Context, msgs ...[]byte) error {
	res := make([]*gpubsub.PublishResult, 0, len(msgs))
	for _, data := range msgs {
		r := p.topic.Publish(ctx, &gpubsub.Message{Data: data})
		res = append(res, r)
	}
	var last error
	for _, r := range res {
		_, err := r.Get(ctx)
		if err != nil {
			last = err
			report.Error(ctx, err)
		}
	}
	return last
}

func (p *gcpPublisher) Batch(ctx context.Context) (Batch, error) {
	return &gcpBatch{topic: p.topic}, nil
}

type gcpBatch struct {
	topic *gpubsub.Topic
	res   []*gpubsub.PublishResult
}

func (b *gcpBatch) Publish(ctx context.Context, msgs ...[]byte) error {
	for _, data := range msgs {
		r := b.topic.Publish(ctx, &gpubsub.Message{Data: data})
		b.res = append(b.res, r)
	}
	return nil
}

func (b *gcpBatch) Flush(ctx context.Context) error {
	var last error
	for _, r := range b.res {
		_, err := r.Get(ctx)
		if err != nil {
			last = err
			report.Error(ctx, err)
		}
	}
	b.res = nil
	return last
}

func (b *gcpBatch) Close() error {
	if len(b.res) != 0 {
		report.Message(context.Background(), "dropped %d events", len(b.res))
	}
	return nil
}

// PublishJSON publishes values as JSON to Pub/Sub topic synchronously.
func PublishJSON(ctx context.Context, p MinPublisher, vals ...interface{}) error {
	msgs := make([][]byte, 0, len(vals))
	for _, v := range vals {
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to encode the value: %v", err)
		}
		msgs = append(msgs, data)
	}
	return p.Publish(ctx, msgs...)
}
