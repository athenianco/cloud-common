package pubsub

import (
	"context"
)

// Message is the payload of a Pub/Sub event.
type Message struct {
	Data  []byte            `json:"data"`
	Attrs map[string]string `json:"attributes,omitempty"`
}

// Handler is a Pub/Sub message handler (subscriber).
type Handler func(ctx context.Context, msg Message) error
