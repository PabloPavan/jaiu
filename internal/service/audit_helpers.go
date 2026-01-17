package service

import (
	"context"

	"github.com/PabloPavan/jaiu/internal/auditctx"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

func recordAuditAttempt(ctx context.Context, repo ports.AuditRepository, actionBase, entityType, entityID string, metadata map[string]any) {
	recordAudit(ctx, repo, actionBase+".attempt", entityType, entityID, metadata)
}

func recordAuditSuccess(ctx context.Context, repo ports.AuditRepository, actionBase, entityType, entityID string, metadata map[string]any) {
	recordAudit(ctx, repo, actionBase+".success", entityType, entityID, metadata)
}

func recordAuditFailure(ctx context.Context, repo ports.AuditRepository, actionBase, entityType, entityID string, metadata map[string]any, err error) {
	meta := copyMetadata(metadata)
	if err != nil {
		meta["error"] = err.Error()
	}
	recordAudit(ctx, repo, actionBase+".failure", entityType, entityID, meta)
}

func recordAudit(ctx context.Context, repo ports.AuditRepository, action, entityType, entityID string, metadata map[string]any) {
	if repo == nil {
		return
	}
	event := buildAuditEvent(ctx, action, entityType, entityID, metadata)
	_ = repo.Record(ctx, event)
}

func buildAuditEvent(ctx context.Context, action, entityType, entityID string, metadata map[string]any) domain.AuditEvent {
	info := auditctx.FromContext(ctx)
	meta := copyMetadata(metadata)
	if info.Actor.Email != "" {
		meta["actor_email"] = info.Actor.Email
	}

	event := domain.AuditEvent{
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   meta,
		IP:         info.Request.IP,
		UserAgent:  info.Request.UserAgent,
	}
	if info.Actor.ID != "" {
		event.ActorID = info.Actor.ID
		event.ActorRole = info.Actor.Role
	}
	return event
}

func copyMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return map[string]any{}
	}
	copied := make(map[string]any, len(metadata))
	for key, value := range metadata {
		copied[key] = value
	}
	return copied
}
