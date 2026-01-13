package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PabloPavan/jaiu/internal/app"
)

func main() {
	cfg := app.Config{
		Addr:          envOr("ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to start: %v", err)
	}
	defer application.Close()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           application.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
