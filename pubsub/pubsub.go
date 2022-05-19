package pubsub

import (
	"context"
	"encoding/json"
	"sync"
)

type MinPublisher interface {
	// PublishMsg publishes messages to the Pub/Sub topic synchronously.
	PublishMsg(ctx context.Context, msgs ...*Message) error
}

type Publisher interface {
	MinPublisher
	// Publish messages to the Pub/Sub topic synchronously.
	Publish(ctx context.Context, msgs ...[]byte) error
	// Batch starts a batch publish operation.
	Batch(ctx context.Context) (Batch, error)
}

type Batch interface {
	// Publish messages to the Pub/Sub topic asynchronously.
	Publish(ctx context.Context, msgs ...[]byte) error
	// PublishMsg publishes messages to the Pub/Sub topic asynchronously.
	PublishMsg(ctx context.Context, msgs ...*Message) error
	// Flush all buffered messages.
	Flush(ctx context.Context) error
	// Close the batch without publishing buffered messages.
	Close() error
}

// NewMemPublisher creates a memory-based publisher implementation that is useful for testing.
func NewMemPublisher() *MemPublisher {
	return &MemPublisher{}
}

var _ Publisher = (*MemPublisher)(nil)

// MemPublisher stores events to memory.
type MemPublisher struct {
	mu     sync.Mutex
	events []Message
}

// GetEvents gets all received events and clears the list.
func (p *MemPublisher) GetEvents() []Message {
	p.mu.Lock()
	events := p.events
	p.events = nil
	p.mu.Unlock()
	return events
}

func (p *MemPublisher) publish(m *Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	var msg Message
	if err = json.Unmarshal(data, &msg); err != nil {
		return err
	}
	p.events = append(p.events, msg)
	return nil
}

// Publish saves events to memory.
func (p *MemPublisher) Publish(ctx context.Context, msgs ...[]byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, m := range msgs {
		if err := p.publish(&Message{Data: m}); err != nil {
			return err
		}
	}
	return nil
}

// PublishMsg saves events to memory.
func (p *MemPublisher) PublishMsg(ctx context.Context, msgs ...*Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, m := range msgs {
		if err := p.publish(m); err != nil {
			return err
		}
	}
	return nil
}

// Batch creates a batch that saves events to memory.
func (p *MemPublisher) Batch(ctx context.Context) (Batch, error) {
	return &memBatch{p: p}, nil
}

type memBatch struct {
	p *MemPublisher
}

func (b *memBatch) Publish(ctx context.Context, msgs ...[]byte) error {
	return b.p.Publish(ctx, msgs...)
}

func (b *memBatch) PublishMsg(ctx context.Context, msgs ...*Message) error {
	return b.p.PublishMsg(ctx, msgs...)
}

func (b *memBatch) Flush(ctx context.Context) error {
	return nil
}

func (b *memBatch) Close() error {
	return nil
}
