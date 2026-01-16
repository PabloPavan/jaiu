package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type httpMetrics struct {
	requests metric.Int64Counter
	duration metric.Float64Histogram
}

var (
	metricsOnce sync.Once
	metrics     httpMetrics
	tracer      trace.Tracer
)

func Observability() func(http.Handler) http.Handler {
	metricsOnce.Do(initHTTPMetrics)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, "http.request", trace.WithSpanKind(trace.SpanKindServer))
			ctx, activity := withUserActivity(ctx)
			r = r.WithContext(ctx)
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}

			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}

			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.route", route),
				attribute.String("http.target", r.URL.Path),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("net.peer.ip", clientIP(r)),
				attribute.Int("http.status_code", status),
			}

			metrics.requests.Add(ctx, 1, metric.WithAttributes(attrs...))
			metrics.duration.Record(ctx, float64(time.Since(start).Milliseconds()), metric.WithAttributes(attrs...))

			logger := observability.Logger(ctx)
			reqID := middleware.GetReqID(ctx)
			if reqID != "" {
				logger = logger.With(slog.String("request_id", reqID))
			}
			logger = logger.With(
				slog.String("method", r.Method),
				slog.String("route", route),
				slog.Int("status", status),
				slog.Int("bytes", recorder.bytes),
				slog.String("remote_ip", clientIP(r)),
				slog.Duration("latency", time.Since(start)),
			)
			if activity != nil && activity.UserID != "" {
				logger = logger.With(
					slog.String("user_id", activity.UserID),
					slog.String("user_role", activity.Role),
				)
				span.SetAttributes(
					attribute.String("enduser.id", activity.UserID),
					attribute.String("enduser.role", activity.Role),
				)
			}

			level := slog.LevelInfo
			if status >= http.StatusInternalServerError {
				level = slog.LevelError
			} else if status >= http.StatusBadRequest {
				level = slog.LevelWarn
			}
			logger.Log(ctx, level, "http request")

			span.SetAttributes(attrs...)
			if status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(status))
			}
			span.SetName(r.Method + " " + route)
			span.End()
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func initHTTPMetrics() {
	meter := otel.Meter("http.server")
	tracer = otel.Tracer("http.server")
	metrics.requests, _ = meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total de requisicoes HTTP"),
	)
	metrics.duration, _ = meter.Float64Histogram(
		"http.server.duration_ms",
		metric.WithDescription("Duracao das requisicoes HTTP em ms"),
	)
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
