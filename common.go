package common

import (
	"context"
	"os"
	"time"

	"github.com/athenianco/cloud-common/report"
)

// EnsureTimeout makes sure there is a timeout set for the context.
func EnsureTimeout(ctx context.Context) (context.Context, func()) {
	// this margin gives us some time to push metrics and other things
	const margin = 5 * time.Second

	if deadline, ok := ctx.Deadline(); ok {
		dt := time.Until(deadline)
		report.Info(ctx, "deadline is set to %v", dt)
		if dt > margin*3 {
			return context.WithDeadline(ctx, deadline.Add(-margin))
		}
		// leave it as-is
		return ctx, func() {}
	}
	const defaultTimeout = time.Minute
	timeout := defaultTimeout
	// will be set in TF file
	if s := os.Getenv("ATHENIAN_TIMEOUT"); s != "" {
		dt, err := time.ParseDuration(s)
		if err != nil {
			report.Error(ctx, err)
		} else {
			timeout = dt
		}
	}
	report.Info(ctx, "deadline is not set, assuming timeout of %v", timeout)
	return context.WithTimeout(ctx, timeout-margin)
}
