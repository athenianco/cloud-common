package pubsub

import (
	"context"
	"sync"
)

type MinPublisher interface {
	// Publish messages to the Pub/Sub topic synchronously.
	Publish(ctx context.Context, msgs ...[]byte) error
}

type Publisher interface {
	// Publish messages to the Pub/Sub topic synchronously.
	Publish(ctx context.Context, msgs ...[]byte) error
	// Batch starts a batch publish operation.
	Batch(ctx context.Context) (Batch, error)
}

type Batch interface {
	// Publish messages to the Pub/Sub topic asynchronously.
	Publish(ctx context.Context, msgs ...[]byte) error
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
	events []string
}

// GetEvents gets all received events and clears the list.
func (p *MemPublisher) GetEvents() []string {
	p.mu.Lock()
	events := p.events
	p.events = nil
	p.mu.Unlock()
	return events
}

// Publish saves events to memory.
func (p *MemPublisher) Publish(ctx context.Context, msgs ...[]byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, m := range msgs {
		p.events = append(p.events, string(m))
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

func (b *memBatch) Flush(ctx context.Context) error {
	return nil
}

func (b *memBatch) Close() error {
	return nil
}
