package report

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

// parseJSONLogs required for comparing the logs without key ordering.
func parseJSONLogs(t testing.TB, r io.Reader) []map[string]interface{} {
	dec := json.NewDecoder(r)
	var arr []map[string]interface{}
	for {
		m := make(map[string]interface{})
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		arr = append(arr, m)
	}
	return arr
}

func TestInfo(t *testing.T) {
	oldOut, oldErr := zlogOut, zlogErr
	defer func() {
		zlogOut, zlogErr = oldOut, oldErr
	}()

	buf := bytes.NewBuffer(nil)
	zlogOut = zerolog.New(buf).Level(zerolog.InfoLevel)
	zlogErr = zlogOut

	ctx := context.Background()
	ctx = WithStringValue(ctx, "foo", "val")

	Debug(ctx, "debug message: %d", 0)
	Info(ctx, "info message: %d", 123)
	Message(ctx, "warn message: %d", 321)

	ctx = WithIntValue(ctx, "bar", 2)
	ctx = WithStringValues(ctx, "baz", []string{"A", "B"})
	Error(ctx, errors.New("error message"))

	require.Equal(t, parseJSONLogs(t, strings.NewReader(`
{"severity":"info","foo":"val","message":"info message: 123"}
{"severity":"warn","foo":"val","message":"warn message: 321"}
{"severity":"error","foo":"val","bar":2,"baz":["A","B"],"message":"error message"}
`)), parseJSONLogs(t, buf))

	Debug(WithDebug(ctx), "debug message: %d", 1)

	require.Equal(t, parseJSONLogs(t, strings.NewReader(`
{"severity":"debug","foo":"val","bar":2,"baz":["A","B"],"message":"debug message: 1"}
`)), parseJSONLogs(t, buf))
}
