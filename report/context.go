package report

import "context"

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

type contextValueKey struct{}

type contextValue struct {
	head  *contextValue
	next  *contextValue
	key   string
	value interface{}
}

type Value struct {
	cur *contextValue
	end *contextValue
}

func (v *Value) Next() *Value {
	if v == nil || v.cur == nil || v.cur.next == nil || v.cur == v.end {
		return nil
	}
	return &Value{
		cur: v.cur.next,
		end: v.end,
	}
}

func (v *Value) Key() string {
	if v == nil || v.cur == nil {
		return ""
	}
	return v.cur.key
}

func (v *Value) Value() interface{} {
	if v == nil || v.cur == nil {
		return ""
	}
	return v.cur.value
}

func getContextValue(ctx context.Context) *contextValue {
	v, _ := ctx.Value(contextValueKey{}).(*contextValue)
	return v
}

func GetContextValues(ctx context.Context) *Value {
	end := getContextValue(ctx)
	if end == nil {
		return nil
	}
	first := end.head
	return &Value{
		cur: first,
		end: end,
	}
}

func GetContextMap(ctx context.Context) map[string]interface{} {
	m := make(map[string]interface{})
	for v := GetContextValues(ctx); v != nil; v = v.Next() {
		m[v.Key()] = v.Value()
	}
	return m
}

func withContextValue(ctx context.Context, key string, val interface{}) context.Context {
	prev := getContextValue(ctx)
	v := &contextValue{key: key, value: val}
	if prev == nil {
		v.head = v
	} else {
		v.head = prev.head
		prev.next = v
	}
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
