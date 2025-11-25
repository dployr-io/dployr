package shared

import (
	"context"
	"testing"
	"time"

	"github.com/dployr-io/dployr/pkg/store"
)

func TestContextEnrichment(t *testing.T) {
	ctx := context.Background()
	enriched := EnrichContext(ctx)

	requestID := RequestID(enriched)
	if requestID == "" {
		t.Error("EnrichContext should add request ID")
	}

	traceID := TraceID(enriched)
	if traceID == "" {
		t.Error("EnrichContext should add trace ID")
	}
}

func TestUserContext(t *testing.T) {
	ctx := context.Background()
	user := &store.User{
		ID:        "user123",
		Email:     "test@example.com",
		Role:      store.RoleDeveloper,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx = context.WithValue(ctx, CtxUserIDKey, user)

	retrieved, err := UserFromContext(ctx)
	if err != nil {
		t.Errorf("UserFromContext() error = %v", err)
	}
	if retrieved.ID != user.ID {
		t.Errorf("Retrieved user ID = %q, want %q", retrieved.ID, user.ID)
	}
}

func TestUserContextMissing(t *testing.T) {
	ctx := context.Background()
	_, err := UserFromContext(ctx)
	if err == nil {
		t.Error("UserFromContext() should error when user is missing")
	}
}

func TestMeta(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequest(ctx, "req123")
	ctx = WithTrace(ctx, "trace456")

	user := &store.User{ID: "user789", Email: "test@example.com"}
	ctx = context.WithValue(ctx, CtxUserIDKey, user)

	meta := Meta(ctx)
	if meta.RequestID != "req123" {
		t.Errorf("Meta.RequestID = %q, want %q", meta.RequestID, "req123")
	}
	if meta.TraceID != "trace456" {
		t.Errorf("Meta.TraceID = %q, want %q", meta.TraceID, "trace456")
	}
	if meta.User.ID != "user789" {
		t.Errorf("Meta.User.ID = %q, want %q", meta.User.ID, "user789")
	}
}
