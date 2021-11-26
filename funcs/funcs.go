package funcs

import (
	"context"
	"net/http"

	"github.com/athenianco/cloud-common/pubsub"
)

// Initializer is an interface for stateful services that require initialization.
// Zero value of the implementation must be usable for calling Init.
type Initializer interface {
	// Init the internal state of the handler.
	// This usually includes reading environment variables,
	Init() error
}

// WebhookHandler is an interface for cloud functions that handle HTTP requests.
type WebhookHandler interface {
	Initializer
	http.Handler
}

// PubSubHandler is an interface for cloud functions that handle Pub/Sub messages.
type PubSubHandler interface {
	Initializer
	// HandleMessage processes a single PubSub message.
	HandleMessage(ctx context.Context, msg *pubsub.Message) error
}
