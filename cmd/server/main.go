package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PabloPavan/jaiu/internal/app"
	"github.com/PabloPavan/jaiu/internal/observability"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	obs, err := observability.Init(ctx, observability.Config{
		ServiceName:    observability.ServiceName("jaiu-api"),
		ServiceVersion: os.Getenv("APP_VERSION"),
		Environment:    os.Getenv("APP_ENV"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
	})
	logger := obs.Logger
	if err != nil {
		logger.Error("failed to initialize observability", "err", err)
	}

	cfg := app.Config{
		Addr:          envOr("ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	}

	application, err := app.New(cfg)
	if err != nil {
		logger.Error("failed to start application", "err", err)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := obs.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown observability", "err", err)
		}
		os.Exit(1)
	}
	defer application.Close()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           application.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server started", slog.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown server", "err", err)
	}
	if err := obs.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown observability", "err", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
