package main

import (
	"context"
	"hash/fnv"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres"
	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/PabloPavan/jaiu/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	obs, err := observability.Init(ctx, observability.Config{
		ServiceName:    observability.ServiceName("jaiu-renewal-worker"),
		ServiceVersion: os.Getenv("APP_VERSION"),
		Environment:    os.Getenv("APP_ENV"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
	})
	logger := obs.Logger
	if err != nil {
		logger.Error("failed to initialize observability", "err", err)
	}

	cfg := config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Hour:        envInt("RENEWAL_HOUR", 0),
		Minute:      envInt("RENEWAL_MINUTE", 5),
	}

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL is required")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown observability", "err", err)
		}
		os.Exit(1)
	}

	pool, err := newPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown observability", "err", err)
		}
		os.Exit(1)
	}
	defer pool.Close()

	job := service.NewRenewalJob(
		postgres.NewSubscriptionRepository(pool),
		postgres.NewPlanRepository(pool),
		postgres.NewBillingPeriodRepository(pool),
		postgres.NewSubscriptionBalanceRepository(pool),
		postgres.NewPaymentTxRunner(pool),
	)

	lockKey := advisoryKey("jaiu:renewal_job")
	tracer := otel.Tracer("renewal-worker")
	meter := otel.Meter("renewal-worker")
	runCounter, _ := meter.Int64Counter("renewal.job.runs", metric.WithDescription("Execucoes do job de renovacao"))
	errorCounter, _ := meter.Int64Counter("renewal.job.errors", metric.WithDescription("Erros do job de renovacao"))
	durationHist, _ := meter.Float64Histogram("renewal.job.duration_ms", metric.WithDescription("Duracao do job de renovacao em ms"))

	go runDaily(ctx, cfg.Hour, cfg.Minute, func(runCtx context.Context) {
		runCtx, span := tracer.Start(runCtx, "renewal.run")
		start := time.Now()
		locked, err := withAdvisoryLock(runCtx, pool, lockKey, job.Run)
		duration := float64(time.Since(start).Milliseconds())
		attrs := []attribute.KeyValue{
			attribute.Bool("lock_acquired", locked),
		}

		runCounter.Add(runCtx, 1, metric.WithAttributes(attrs...))
		durationHist.Record(runCtx, duration, metric.WithAttributes(attrs...))
		span.SetAttributes(attrs...)

		if err != nil {
			errorCounter.Add(runCtx, 1, metric.WithAttributes(attrs...))
			observability.Logger(runCtx).Error("renewal job failed", "err", err, "lock_acquired", locked)
			span.RecordError(err)
			span.SetStatus(codes.Error, "renewal job failed")
		} else if !locked {
			observability.Logger(runCtx).Debug("renewal job skipped", "lock_acquired", false)
		}
		span.End()
	})

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := obs.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown observability", "err", err)
	}
}

type config struct {
	DatabaseURL string
	Hour        int
	Minute      int
}

func newPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return postgres.NewPool(ctx, url)
}

func runDaily(ctx context.Context, hour, minute int, job func(context.Context)) {
	for {
		next := nextRun(time.Now(), hour, minute)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			job(ctx)
		}
	}
}

func nextRun(now time.Time, hour, minute int) time.Time {
	loc := now.Location()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func withAdvisoryLock(ctx context.Context, pool *pgxpool.Pool, key int64, job func(context.Context) error) (bool, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		return false, err
	}
	if !locked {
		return false, nil
	}
	defer func() {
		_, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", key)
	}()

	return true, job(ctx)
}

func advisoryKey(value string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(value))
	return int64(hash.Sum64())
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
