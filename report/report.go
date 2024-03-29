package report

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/athenianco/cloud-common/gcp"
)

var (
	zlogOut = zerolog.New(os.Stdout).With().Timestamp().Logger()
	zlogErr = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

func init() {
	// this is what GCP expects
	zerolog.LevelFieldName = "severity"
	zerolog.MessageFieldName = "message"
	zerolog.ErrorFieldName = zerolog.MessageFieldName
	zerolog.TimeFieldFormat = time.RFC3339Nano

	lvl := zerolog.InfoLevel
	if warn := os.Getenv("ATHENIAN_COMMON_WARN"); warn == "true" {
		lvl = zerolog.WarnLevel
	}
	if debug := os.Getenv("ATHENIAN_COMMON_DEBUG"); debug == "true" {
		lvl = zerolog.DebugLevel
	}
	zlogOut = zlogOut.Level(lvl)
	zlogErr = zlogErr.Level(lvl)
	if gcp.IsCloudFunction() {
		// redirect error log to stdout as well
		zlogErr = zlogOut
	}
}

// IgnoredError is an error that should be ignored by error reporting services.
type IgnoredError interface {
	error
	Ignored() bool
}

type ignoredError struct {
	error
}

// NewIgnoredError creates a generic error that should be ignored by error reporting services.
func NewIgnoredError(err error) IgnoredError {
	return &ignoredError{error: err}
}

func (e *ignoredError) Ignored() bool {
	return true
}

func (e *ignoredError) Unwrap() error {
	return e.error
}

const (
	instanceIDLen = 8
)

var (
	global     Reporter = reporter{}
	instanceID          = genID()
)

func genID() string {
	var b [instanceIDLen]byte
	n, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	} else if n != len(b) {
		panic("short read from rand")
	}
	return hex.EncodeToString(b[:])
}

// InstanceID returns a unique application instance ID.
func InstanceID() string {
	return instanceID
}

func Global() Reporter {
	return global
}

func SetReporter(r Reporter) {
	global = r
}

type EventID string

type Reporter interface {
	CaptureDebug(ctx context.Context, format string, args ...interface{})
	CaptureInfo(ctx context.Context, format string, args ...interface{})
	CaptureMessage(ctx context.Context, format string, args ...interface{})
	CaptureError(ctx context.Context, err error) EventID
}

func Debug(ctx context.Context, format string, args ...interface{}) {
	if global == nil {
		return
	}
	countReportLogsInc(ctx, severityDebug)
	global.CaptureDebug(ctx, format, args...)
}

func Info(ctx context.Context, format string, args ...interface{}) {
	if global == nil {
		return
	}
	countReportLogsInc(ctx, severityInfo)
	global.CaptureInfo(ctx, format, args...)
}

func Message(ctx context.Context, format string, args ...interface{}) {
	if global == nil {
		return
	}
	countReportLogsInc(ctx, severityWarning)
	global.CaptureMessage(ctx, format, args...)
}

func Error(ctx context.Context, err error) EventID {
	if global == nil || err == nil {
		return ""
	}
	switch err := err.(type) {
	case interface {
		Temporary() bool
	}:
		if err.Temporary() {
			return ""
		}
	case IgnoredError:
		if err.Ignored() {
			return ""
		}
	}
	countReportLogsInc(ctx, severityError)
	return global.CaptureError(ctx, err)
}

var finalizers []func(ctx context.Context) error

// RegisterFlusher registers a flush function that must be called to send monitoring information.
func RegisterFlusher(f func(ctx context.Context) error) {
	finalizers = append(finalizers, f)
}

// Flush must be called to ensure all reports and metrics were sent to the monitoring service(s).
func Flush(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var (
		wg   sync.WaitGroup
		errc = make(chan error, len(finalizers))
	)
	for _, f := range finalizers {
		f := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f(ctx); err != nil {
				zlogErr.Error().Err(err).Send()
				errc <- err
			}
		}()
	}
	// We intentionally wait only for the first event: either all finalizers are done, or timeout was hit.
	// This function must not block longer than the allowed timeout, so keep the finalizer hanging in the background.
	go func() {
		wg.Wait()
		cancel()
	}()
	<-ctx.Done()
	select {
	case err := <-errc:
		return err
	default:
	}
	return nil
}

func Default() Reporter {
	return reporter{}
}

type reporter struct{}

func (reporter) fromContext(ctx context.Context, ev *zerolog.Event) *zerolog.Event {
	for key, val := range GetContextMap(ctx) {
		ev = ev.Interface(key, val)
	}
	return ev
}

func (r reporter) CaptureDebug(ctx context.Context, format string, args ...interface{}) {
	l := zlogOut
	if GetDebug(ctx) {
		l = l.Level(zerolog.DebugLevel)
	}
	r.fromContext(ctx, l.Debug()).Msgf(format, args...)
}

func (r reporter) CaptureInfo(ctx context.Context, format string, args ...interface{}) {
	r.fromContext(ctx, zlogOut.Info()).Msgf(format, args...)
}

func (r reporter) CaptureMessage(ctx context.Context, format string, args ...interface{}) {
	r.fromContext(ctx, zlogOut.Warn()).Msgf(format, args...)
}

func (r reporter) CaptureError(ctx context.Context, err error) EventID {
	r.fromContext(ctx, zlogErr.Error()).Err(err).Send()
	return ""
}
