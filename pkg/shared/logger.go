package shared

import (
	"context"
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
    return slog.New(handler)
}

func LogWithContext(ctx context.Context) *slog.Logger {
    requestID, _ := RequestFromContext(ctx)
    traceID, _ := TraceFromContext(ctx)
    user, _ := UserFromContext(ctx)
    project, _ := ProjectFromContext(ctx)

    attrs := []slog.Attr{
        slog.String("request_id", safeString(&requestID)),
        slog.String("trace_id", safeString(&traceID)),
        slog.String("user_id", safeString(&user.ID)),
        slog.String("project_id", safeString(&project.ID)),
    }

    args := make([]any, len(attrs))
    for i, a := range attrs {
        args[i] = a
    }

    return slog.With(args...)
}

func safeString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
