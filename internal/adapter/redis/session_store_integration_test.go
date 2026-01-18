//go:build integration

package redis

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	redislib "github.com/redis/go-redis/v9"
)

func integrationRedisClient(t *testing.T) *redislib.Client {
	t.Helper()

	if os.Getenv("JAIU_INTEGRATION") != "true" {
		t.Skip("JAIU_INTEGRATION is not set to true")
	}

	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = os.Getenv("REDIS_ADDR")
	}
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR or REDIS_ADDR is required for integration tests")
	}

	db := 0
	if raw := os.Getenv("TEST_REDIS_DB"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			t.Fatalf("invalid TEST_REDIS_DB: %v", err)
		}
		db = parsed
	}

	client := redislib.NewClient(&redislib.Options{
		Addr:     addr,
		Password: os.Getenv("TEST_REDIS_PASSWORD"),
		DB:       db,
	})
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}

	return client
}

// Testa o ciclo de vida completo da sessao no Redis.
func TestSessionStoreIntegration(t *testing.T) {
	client := integrationRedisClient(t)
	store := NewSessionStore(client)
	ctx := context.Background()

	session := ports.Session{
		UserID:    "user-1",
		Role:      domain.RoleAdmin,
		ExpiresAt: time.Now().Add(1 * time.Minute),
	}
	token, err := store.Create(ctx, session)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	loaded, err := store.Get(ctx, token)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if loaded.UserID != session.UserID {
		t.Fatalf("expected user %q, got %q", session.UserID, loaded.UserID)
	}

	if err := store.Delete(ctx, token); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := store.Get(ctx, token); err != ports.ErrNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

// Testa que sessao expirada retorna not found.
func TestSessionStoreIntegrationExpired(t *testing.T) {
	client := integrationRedisClient(t)
	store := NewSessionStore(client)
	ctx := context.Background()

	expired := ports.Session{
		UserID:    "user-2",
		Role:      domain.RoleAdmin,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	token, err := store.Create(ctx, expired)
	if err != nil {
		t.Fatalf("create expired session: %v", err)
	}
	if _, err := store.Get(ctx, token); err != ports.ErrNotFound {
		t.Fatalf("expected not found for expired session, got %v", err)
	}
}

// Testa Get com token vazio retornando not found.
func TestSessionStoreIntegrationEmptyToken(t *testing.T) {
	client := integrationRedisClient(t)
	store := NewSessionStore(client)
	ctx := context.Background()

	if _, err := store.Get(ctx, ""); err != ports.ErrNotFound {
		t.Fatalf("expected not found for empty token, got %v", err)
	}
}
