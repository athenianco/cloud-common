package funcs

import (
	"context"
	"encoding/json"
	"net/http"

	common "github.com/athenianco/cloud-common"
	"github.com/athenianco/cloud-common/pubsub"
	"github.com/athenianco/cloud-common/report"
	"github.com/athenianco/cloud-common/report/sentry"
	"github.com/athenianco/cloud-common/service"
)

const port = ":8080"

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

func RunHTTP(h WebhookHandler) bool {
	ctx := context.Background()

	defer report.Flush()
	defer sentry.RecoverAndPanic(ctx)

	if err := h.Init(); err != nil {
		panic(err)
	}
	enabled, err := service.Register(ctx)
	if err != nil {
		panic(err)
	}
	if !enabled {
		report.Info(ctx, "service disabled")
		return false
	}

	http.Handle("/", h)
	if err := http.ListenAndServe(port, nil); err != nil {
		panic(err)
	}
	return true
}

type pubsubHandler struct {
	PubSubHandler
}

func (h *pubsubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// https://github.com/GoogleCloudPlatform/golang-samples/blob/31bb00e8dd7407c229442f37fb8b99d24df15233/eventarc/pubsub/main.go#L31
	var msg struct {
		pubsub.Message `json:"message"`
	}
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		handleErr(ctx, w, err, http.StatusBadRequest)
		return
	}
	ctx, cancel := common.EnsureTimeout(ctx)
	defer cancel()
	if err := h.HandleMessage(ctx, &msg.Message); err != nil {
		// TODO: better status codes
		handleErr(ctx, w, err, http.StatusInternalServerError)
		return
	}
}

func RunPubSub(h PubSubHandler) bool {
	return RunHTTP(&pubsubHandler{h})
}

func handleErr(ctx context.Context, w http.ResponseWriter, err error, status int) {
	report.Error(ctx, err)
	http.Error(w, err.Error(), status)
}
