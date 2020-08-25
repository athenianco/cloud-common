package sentry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/athenianco/cloud-common/report"
	"github.com/getsentry/sentry-go"
)

func init() {
	if os.Getenv("SENTRY_DSN") == "" {
		return
	}
	tr := sentry.NewHTTPSyncTransport()
	tr.Timeout = 5 * time.Second
	if err := sentry.Init(sentry.ClientOptions{
		Transport:  tr,
		ServerName: os.Getenv("SENTRY_SERVER"),
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
	scope.SetUser(sentry.User{ID: report.GetUserID(ctx)})
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

func (r reporter) CaptureError(ctx context.Context, err error) {
	r.r.CaptureError(ctx, err)
	h := hubFromContext(ctx)
	h.WithScope(func(scope *sentry.Scope) {
		setScope(ctx, scope)
		h.CaptureException(err)
	})
}

func RecoverAndPanic(ctx context.Context) {
	if r := recover(); r != nil {
		h := hubFromContext(ctx)
		h.WithScope(func(scope *sentry.Scope) {
			setScope(ctx, scope)
			h.Recover(r)
		})
		panic(r)
	}
}
