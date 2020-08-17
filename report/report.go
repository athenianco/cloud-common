package report

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug := os.Getenv("ATHENIAN_COMMON_DEBUG"); debug == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
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

type Reporter interface {
	CaptureInfo(ctx context.Context, format string, args ...interface{})
	CaptureMessage(ctx context.Context, format string, args ...interface{})
	CaptureError(ctx context.Context, err error)
}

func Info(ctx context.Context, format string, args ...interface{}) {
	if global == nil {
		return
	}
	global.CaptureInfo(ctx, format, args...)
}

func Message(ctx context.Context, format string, args ...interface{}) {
	if global == nil {
		return
	}
	global.CaptureMessage(ctx, format, args...)
}

func Error(ctx context.Context, err error) {
	if global == nil || err == nil {
		return
	}
	switch err := err.(type) {
	case interface {
		Temporary() bool
	}:
		if err.Temporary() {
			return
		}
	}
	global.CaptureError(ctx, err)
}

var finalizers []func() error

// RegisterFlusher registers a flush function that must be called to send monitoring information.
func RegisterFlusher(f func() error) {
	finalizers = append(finalizers, f)
}

// Flush must be called to ensure all reports and metrics were sent to the monitoring service(s).
func Flush() error {
	var last error
	for _, f := range finalizers {
		if err := f(); err != nil {
			log.Error().Err(err).Send()
			last = err
		}
	}
	return last
}

func Default() Reporter {
	return reporter{}
}

type reporter struct{}

func (reporter) fromContext(ctx context.Context, ev *zerolog.Event) *zerolog.Event {
	for v := GetContextValues(ctx); v != nil; v = v.Next() {
		key, val := v.Key(), v.Value()
		ev = ev.Interface(key, val)
	}
	return ev
}

func (r reporter) CaptureInfo(ctx context.Context, format string, args ...interface{}) {
	r.fromContext(ctx, log.Info()).Msgf(format, args...)
}

func (r reporter) CaptureMessage(ctx context.Context, format string, args ...interface{}) {
	r.fromContext(ctx, log.Info()).Msgf(format, args...)
}

func (r reporter) CaptureError(ctx context.Context, err error) {
	r.fromContext(ctx, log.Error()).Err(err).Send()
}
