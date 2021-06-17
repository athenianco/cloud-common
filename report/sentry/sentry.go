package sentry

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/athenianco/cloud-common/report"
)

type EventErrFunc func(ctx context.Context, ev *sentry.Event, err error)

var (
	beforeMu   sync.RWMutex
	beforeSend []EventErrFunc
)

func RegisterBeforeSend(fnc EventErrFunc) {
	beforeMu.Lock()
	defer beforeMu.Unlock()
	beforeSend = append(beforeSend, fnc)
}

func getSendHooks() []EventErrFunc {
	beforeMu.RLock()
	funcs := beforeSend
	beforeMu.RUnlock()
	return funcs
}

func init() {
	if os.Getenv("SENTRY_DSN") == "" {
		return
	}
	tr := sentry.NewHTTPSyncTransport()
	tr.Timeout = 5 * time.Second
	if err := sentry.Init(sentry.ClientOptions{
		Transport:  tr,
		ServerName: os.Getenv("SENTRY_SERVER"),
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if len(event.Fingerprint) == 0 {
				event.Fingerprint = []string{"{{ default }}"}
			}
			err := hint.OriginalException
			// pgconn.PgError
			if e, ok := err.(interface {
				error
				SQLState() string
			}); ok {
				event.Fingerprint = append(event.Fingerprint, e.Error(), e.SQLState())
			}
			for _, fnc := range getSendHooks() {
				fnc(hint.Context, event, err)
			}
			return event
		},
	}); err != nil {
		panic(err)
	}
	report.SetReporter(reporter{r: report.Default()})
	report.RegisterFlusher(func() error {
		sentry.Flush(tr.Timeout)
		return nil
	})
}

var _ report.Reporter = reporter{}

func hubFromContext(ctx context.Context) *sentry.Hub {
	if h := sentry.GetHubFromContext(ctx); h != nil {
		return h
	}
	return sentry.CurrentHub()
}

func setScope(ctx context.Context, scope *sentry.Scope) {
	scope.SetUser(sentry.User{
		ID:       report.GetUserID(ctx),
		Username: report.GetUserName(ctx),
		Email:    report.GetUserEmail(ctx),
	})
}

type reporter struct {
	r report.Reporter
}

func (r reporter) CaptureInfo(ctx context.Context, format string, args ...interface{}) {
	r.r.CaptureInfo(ctx, format, args...)
}

func (r reporter) CaptureMessage(ctx context.Context, format string, args ...interface{}) {
	r.r.CaptureMessage(ctx, format, args...)
	h := hubFromContext(ctx)
	h.WithScope(func(scope *sentry.Scope) {
		setScope(ctx, scope)
		h.CaptureMessage(fmt.Sprintf(format, args...))
	})
}

func (r reporter) CaptureError(ctx context.Context, err error) report.EventID {
	id := r.r.CaptureError(ctx, err)
	h := hubFromContext(ctx)

	h.WithScope(func(scope *sentry.Scope) {
		setScope(ctx, scope)
		id = report.EventID(*h.CaptureException(err))
	})
	return id
}

func CaptureAndPanic(ctx context.Context, r interface{}) {
	h := hubFromContext(ctx)
	h.WithScope(func(scope *sentry.Scope) {
		setScope(ctx, scope)
		h.Recover(r)
	})
	panic(r)
}

func RecoverAndPanic(ctx context.Context) {
	if r := recover(); r != nil {
		CaptureAndPanic(ctx, r)
	}
}
