package report

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func listValues(ctx context.Context) map[string]interface{} {
	m := make(map[string]interface{})
	for v := GetContextValues(ctx); v != nil; v = v.Next() {
		m[v.Key()] = v.Value()
	}
	return m
}

type M = map[string]interface{}

func TestWithValues(t *testing.T) {
	ctx := context.Background()
	require.Equal(t, M{}, listValues(ctx))

	ctx1 := WithStringValue(ctx, "k1", "v1")
	ctx1 = WithStringValue(ctx1, "k2", "v2")
	ctx1 = WithStringValue(ctx1, "k3", "v3")

	ctx2 := WithStringValue(ctx1, "k1", "v1'")
	ctx2 = WithStringValue(ctx2, "k3", "v3'")

	require.Equal(t, M{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}, listValues(ctx1))
	require.Equal(t, M{
		"k1": "v1'",
		"k2": "v2",
		"k3": "v3'",
	}, listValues(ctx2))
}
