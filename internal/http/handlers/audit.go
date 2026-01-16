package handlers

import (
	"net"
	"net/http"

	"github.com/PabloPavan/jaiu/internal/domain"
	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
	"github.com/PabloPavan/jaiu/internal/observability"
)

func (h *Handler) recordAudit(r *http.Request, action, entityType, entityID string, metadata map[string]any) {
	if h.services.Audit == nil {
		return
	}
	event := buildAuditEvent(r, action, entityType, entityID, metadata)
	h.recordAuditEvent(r, event)
}

func (h *Handler) recordAuditWithActor(r *http.Request, actorID, actorRole, action, entityType, entityID string, metadata map[string]any) {
	if h.services.Audit == nil {
		return
	}
	event := buildAuditEvent(r, action, entityType, entityID, metadata)
	event.ActorID = actorID
	event.ActorRole = actorRole
	h.recordAuditEvent(r, event)
}

func (h *Handler) recordAuditEvent(r *http.Request, event domain.AuditEvent) {
	if h.services.Audit == nil {
		return
	}
	if err := h.services.Audit.Record(r.Context(), event); err != nil {
		observability.Logger(r.Context()).Error("audit record failed", "err", err, "action", event.Action, "entity_type", event.EntityType, "entity_id", event.EntityID)
	}
}

func buildAuditEvent(r *http.Request, action, entityType, entityID string, metadata map[string]any) domain.AuditEvent {
	event := domain.AuditEvent{
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadata,
		IP:         requestIP(r),
		UserAgent:  r.UserAgent(),
	}
	if session, ok := httpmw.SessionFromContext(r.Context()); ok {
		event.ActorID = session.UserID
		event.ActorRole = string(session.Role)
	}
	return event
}

func requestIP(r *http.Request) string {
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
