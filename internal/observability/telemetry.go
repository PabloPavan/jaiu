package observability

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.27.0"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	LogLevel       string
}

type Setup struct {
	Logger   *slog.Logger
	Shutdown func(context.Context) error
}

func Init(ctx context.Context, cfg Config) (*Setup, error) {
	setup := &Setup{
		Logger:   slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
		Shutdown: func(context.Context) error { return nil },
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "jaiu"
	}

	if !otlpConfigured() {
		logger := newLogger(cfg, false)
		setup.Logger = logger
		slog.SetDefault(logger)
		return setup, nil
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return setup, err
	}

	var shutdowns []func(context.Context) error

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return setup, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	shutdowns = append(shutdowns, tp.Shutdown)

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return setup, err
	}
	reader := sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(15*time.Second))
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	otel.SetMeterProvider(mp)
	shutdowns = append(shutdowns, mp.Shutdown)

	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return setup, err
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	global.SetLoggerProvider(lp)
	shutdowns = append(shutdowns, lp.Shutdown)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger := newLogger(cfg, true)
	setup.Logger = logger
	setup.Shutdown = func(ctx context.Context) error {
		var shutdownErr error
		for _, fn := range shutdowns {
			if err := fn(ctx); err != nil && shutdownErr == nil {
				shutdownErr = err
			}
		}
		return shutdownErr
	}

	slog.SetDefault(logger)
	return setup, nil
}

func ServiceName(fallback string) string {
	if value := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")); value != "" {
		return value
	}
	return fallback
}

func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(cfg.ServiceVersion))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentNameKey.String(cfg.Environment))
	}

	return resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attrs...),
	)
}

func newLogger(cfg Config, enableOtel bool) *slog.Logger {
	level := parseLevel(cfg.LogLevel)
	stdout := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	handler := slog.Handler(stdout)
	if enableOtel {
		otelHandler := otelslog.NewHandler(cfg.ServiceName)
		handler = newMultiHandler(stdout, otelHandler)
	}
	logger := slog.New(handler).With("service", cfg.ServiceName)
	return logger
}

func parseLevel(raw string) slog.Level {
	level := strings.ToLower(strings.TrimSpace(raw))
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func otlpConfigured() bool {
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")) != "" {
		return true
	}
	return false
}

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	valid := make([]slog.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if handler != nil {
			valid = append(valid, handler)
		}
	}
	return multiHandler{handlers: valid}
}

func (h multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var handleErr error
	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record.Clone()); err != nil && handleErr == nil {
			handleErr = err
		}
	}
	return handleErr
}

func (h multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		next = append(next, handler.WithAttrs(attrs))
	}
	return multiHandler{handlers: next}
}

func (h multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		next = append(next, handler.WithGroup(name))
	}
	return multiHandler{handlers: next}
}
