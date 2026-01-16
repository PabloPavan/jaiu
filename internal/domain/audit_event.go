package domain

import "time"

type AuditEvent struct {
	ID         string
	ActorID    string
	ActorRole  string
	Action     string
	EntityType string
	EntityID   string
	Metadata   map[string]any
	IP         string
	UserAgent  string
	CreatedAt  time.Time
}
