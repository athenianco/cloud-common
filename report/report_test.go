package report

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestWrapError(t *testing.T) {
	orig := errors.New("test")
	err := Errorf("sub: %w", orig)
	require.Equal(t, "sub: test", err.Error())
}

func TestInfo(t *testing.T) {
	oldOut, oldErr := zlogOut, zlogErr
	defer func() {
		zlogOut, zlogErr = oldOut, oldErr
	}()

	buf := bytes.NewBuffer(nil)
	zlogOut = zerolog.New(buf)
	zlogErr = zlogOut

	ctx := context.Background()
	ctx = WithStringValue(ctx, "foo", "val")

	Info(ctx, "debug message: %d", 123)
	Message(ctx, "info message: %d", 321)

	ctx = WithIntValue(ctx, "bar", 2)
	ctx = WithStringValues(ctx, "baz", []string{"A", "B"})
	Error(ctx, errors.New("error message"))

	require.Equal(t, strings.TrimSpace(`
{"severity":"info","foo":"val","message":"debug message: 123"}
{"severity":"warn","foo":"val","message":"info message: 321"}
{"severity":"error","foo":"val","bar":2,"baz":["A","B"],"message":"error message"}
`), strings.TrimSpace(buf.String()))
}
