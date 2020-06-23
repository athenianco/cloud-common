package report

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
)

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

type userIDKey struct{}

// WithUserID attaches a user ID to the context.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// GetUserID returns a user ID to the context, if any.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey{}).(string)
	return id
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
			log.Println(err)
			last = err
		}
	}
	return last
}

type reporter struct{}

func (reporter) CaptureInfo(ctx context.Context, format string, args ...interface{}) {
	// log.Printf(format, args...)
}

func (reporter) CaptureMessage(ctx context.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (reporter) CaptureError(ctx context.Context, err error) {
	log.Println("error:", err)
}
