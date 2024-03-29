package report

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type debugKey struct{}

// WithDebug enables debug logging on this context.
func WithDebug(ctx context.Context) context.Context {
	return context.WithValue(ctx, debugKey{}, true)
}

// GetDebug checks if debug flag was set on the context.
func GetDebug(ctx context.Context) bool {
	v, _ := ctx.Value(debugKey{}).(bool)
	return v
}

type userIDKey struct{}
type userNameKey struct{}
type userEmailKey struct{}

// WithUserID attaches a user ID to the context.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// WithUserName attaches a username to the context.
func WithUserName(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userNameKey{}, id)
}

// WithUserEmail attaches a user email to the context.
func WithUserEmail(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userEmailKey{}, id)
}

// GetUserID returns a user ID to the context, if any.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey{}).(string)
	return id
}

// GetUserName returns a username to the context, if any.
func GetUserName(ctx context.Context) string {
	id, _ := ctx.Value(userNameKey{}).(string)
	return id
}

// GetUserEmail returns a user email to the context, if any.
func GetUserEmail(ctx context.Context) string {
	id, _ := ctx.Value(userEmailKey{}).(string)
	return id
}

type contextValueKey struct{}

type Value struct {
	prev  *Value
	key   string
	value interface{}
}

func (v *Value) Prev() *Value {
	if v == nil || v.prev == nil {
		return nil
	}
	return v.prev
}

func (v *Value) Key() string {
	if v == nil {
		return ""
	}
	return v.key
}

func (v *Value) Value() interface{} {
	if v == nil {
		return nil
	}
	return v.value
}

func GetContextValues(ctx context.Context) *Value {
	v, _ := ctx.Value(contextValueKey{}).(*Value)
	return v
}

func GetContextMap(ctx context.Context) map[string]interface{} {
	m := make(map[string]interface{})
	for v := GetContextValues(ctx); v != nil; v = v.Prev() {
		key := v.Key()
		if _, ok := m[key]; ok {
			continue
		}
		m[key] = v.Value()
	}
	return m
}

func withContextValue(ctx context.Context, key string, val interface{}) context.Context {
	prev := GetContextValues(ctx)
	v := &Value{key: key, value: val, prev: prev}
	return context.WithValue(ctx, contextValueKey{}, v)
}

func WithStringValue(ctx context.Context, key, val string) context.Context {
	return withContextValue(ctx, key, val)
}

func WithIntValue(ctx context.Context, key string, val int) context.Context {
	return WithInt64Value(ctx, key, int64(val))
}

func WithInt64Value(ctx context.Context, key string, val int64) context.Context {
	return withContextValue(ctx, key, val)
}

func WithStringValues(ctx context.Context, key string, val []string) context.Context {
	return withContextValue(ctx, key, val)
}

func WithRandomID(ctx context.Context, key string) context.Context {
	var buf [16]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		panic(err)
	}
	return withContextValue(ctx, key, hex.EncodeToString(buf[:]))
}
