package report

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	old := log.Logger
	defer func() {
		log.Logger = old
	}()

	buf := bytes.NewBuffer(nil)
	log.Logger = zerolog.New(buf)

	ctx := context.Background()
	ctx = WithStringValue(ctx, "foo", "val")

	Info(ctx, "debug message: %d", 123)
	Message(ctx, "info message: %d", 321)

	ctx = WithIntValue(ctx, "bar", 2)
	ctx = WithStringValues(ctx, "baz", []string{"A", "B"})
	Error(ctx, errors.New("error message"))

	require.Equal(t, strings.TrimSpace(`
{"level":"info","foo":"val","message":"debug message: 123"}
{"level":"info","foo":"val","message":"info message: 321"}
{"level":"error","foo":"val","bar":2,"baz":["A","B"],"error":"error message"}
`), strings.TrimSpace(buf.String()))
}
