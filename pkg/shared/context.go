package shared

import (
	"context"
	"dployr/pkg/store"
	"errors"

	"github.com/oklog/ulid/v2"
)

type ContextKey string

const (
	CtxUserIDKey     ContextKey = "user_id"
	CtxRequestIDKey  ContextKey = "request_id"
	CtxTraceIDKey    ContextKey = "trace_id"
)

func WithUser(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxUserIDKey, id)
}

func UserFromContext(ctx context.Context) (*store.User, error) {
    user, ok := ctx.Value(CtxUserIDKey).(*store.User)
    if !ok || user == nil {
        return nil, errors.New("no authenticated user in context")
    }
    return user, nil
}

func WithRequest(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxRequestIDKey, id)
}

func RequestFromContext(ctx context.Context) (string, error) {
	requestID, ok := ctx.Value(CtxRequestIDKey).(string)
	_, err := ulid.Parse(requestID)
    if !ok || err != nil {
        return "", errors.New("no request id in context")
    }
    return requestID, nil
}

func WithTrace(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxTraceIDKey, id)
}

func TraceFromContext(ctx context.Context) (string, error) {
	traceID, ok := ctx.Value(CtxTraceIDKey).(string)
	_, err := ulid.Parse(traceID)
    if !ok || err != nil {
        return "", errors.New("no trace id in context")
    }
    return traceID, nil
}

func User(ctx context.Context) *store.User {
	v, _ := UserFromContext(ctx)
	return v
}


func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(CtxRequestIDKey).(string); ok {
		return v
	}
	return ""
}

func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(CtxTraceIDKey).(string); ok {
		return v
	}
	return ""
}

func EnrichContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(CtxRequestIDKey).(string); !ok {
		ctx = WithRequest(ctx, ulid.Make().String())
	}
	if _, ok := ctx.Value(CtxTraceIDKey).(string); !ok {
		ctx = WithTrace(ctx, ulid.Make().String())
	}
	return ctx
}

type ContextMeta struct {
	RequestID string
	TraceID   string
	User      *store.User
}

func Meta(ctx context.Context) *ContextMeta {
	return &ContextMeta{
		RequestID: RequestID(ctx),
		TraceID:   TraceID(ctx),
		User:      User(ctx),
	}
}
