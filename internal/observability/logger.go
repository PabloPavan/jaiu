package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

func Logger(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	if ctx == nil {
		return logger
	}
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() {
		return logger
	}
	return logger.With(
		slog.String("trace_id", span.TraceID().String()),
		slog.String("span_id", span.SpanID().String()),
	)
}
