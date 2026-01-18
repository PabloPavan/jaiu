package service

import (
	"context"
	"testing"

	"github.com/PabloPavan/jaiu/internal/auditctx"
)

// Testa que copyMetadata cria uma copia independente do mapa original.
func TestCopyMetadata(t *testing.T) {
	original := map[string]any{"key": "value"}
	copied := copyMetadata(original)
	copied["key"] = "changed"

	if original["key"] == "changed" {
		t.Fatal("expected copy to not mutate original map")
	}
	if len(copyMetadata(nil)) != 0 {
		t.Fatal("expected empty map for nil metadata")
	}
}

// Testa a composicao do evento de auditoria com dados do contexto.
func TestBuildAuditEvent(t *testing.T) {
	ctx := context.Background()
	ctx = auditctx.WithRequest(ctx, auditctx.RequestInfo{
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	})
	ctx = auditctx.WithActor(ctx, auditctx.Actor{
		ID:    "user-1",
		Role:  "admin",
		Email: "user@example.com",
	})

	event := buildAuditEvent(ctx, "action", "entity", "entity-1", map[string]any{
		"hello": "world",
	})

	if event.Action != "action" || event.EntityType != "entity" || event.EntityID != "entity-1" {
		t.Fatalf("unexpected event identifiers: %#v", event)
	}
	if event.ActorID != "user-1" || event.ActorRole != "admin" {
		t.Fatalf("unexpected actor info: %#v", event)
	}
	if event.IP != "127.0.0.1" || event.UserAgent != "test-agent" {
		t.Fatalf("unexpected request info: %#v", event)
	}
	if event.Metadata["actor_email"] != "user@example.com" {
		t.Fatalf("expected actor_email in metadata, got %#v", event.Metadata)
	}
}

// Testa que recordAuditFailure registra o erro na metadata.
func TestRecordAuditFailure(t *testing.T) {
	repo := &auditRepoFake{}
	recordAuditFailure(context.Background(), repo, "entity.create", "entity", "id-1", map[string]any{}, context.Canceled)

	if len(repo.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(repo.events))
	}
	event := repo.events[0]
	if event.Action != "entity.create.failure" {
		t.Fatalf("expected failure action, got %q", event.Action)
	}
	if event.Metadata["error"] == "" {
		t.Fatalf("expected error in metadata, got %#v", event.Metadata)
	}
}
