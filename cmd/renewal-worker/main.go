package main

import (
	"context"
	"hash/fnv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres"
	"github.com/PabloPavan/jaiu/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Hour:        envInt("RENEWAL_HOUR", 0),
		Minute:      envInt("RENEWAL_MINUTE", 5),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := newPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer pool.Close()

	job := service.NewRenewalJob(
		postgres.NewSubscriptionRepository(pool),
		postgres.NewPlanRepository(pool),
		postgres.NewBillingPeriodRepository(pool),
		postgres.NewSubscriptionBalanceRepository(pool),
	)

	lockKey := advisoryKey("jaiu:renewal_job")
	go runDaily(ctx, cfg.Hour, cfg.Minute, func(runCtx context.Context) {
		if err := withAdvisoryLock(runCtx, pool, lockKey, job.Run); err != nil {
			log.Printf("renewal job error: %v", err)
		}
	})

	<-ctx.Done()
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

func withAdvisoryLock(ctx context.Context, pool *pgxpool.Pool, key int64, job func(context.Context) error) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer func() {
		_, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", key)
	}()

	return job(ctx)
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
