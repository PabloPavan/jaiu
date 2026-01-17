package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	exec auditExecer
}

type auditExecer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{exec: pool}
}

func NewAuditRepositoryWithTx(tx pgx.Tx) *AuditRepository {
	return &AuditRepository{exec: tx}
}

func (r *AuditRepository) Record(ctx context.Context, event domain.AuditEvent) error {
	if r == nil || r.exec == nil {
		return errors.New("audit repository unavailable")
	}

	actorID, err := stringToUUID(event.ActorID)
	if err != nil {
		return err
	}
	entityID, err := stringToUUID(event.EntityID)
	if err != nil {
		return err
	}

	metadataBytes := []byte("{}")
	if event.Metadata != nil {
		metadataBytes, err = json.Marshal(event.Metadata)
		if err != nil {
			return err
		}
	}
	_, err = r.exec.Exec(ctx, `
		INSERT INTO audit_events (
			actor_id,
			actor_role,
			action,
			entity_type,
			entity_id,
			metadata,
			ip,
			user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, actorID, textTo(event.ActorRole), event.Action, event.EntityType, entityID, metadataBytes, textTo(event.IP), textTo(event.UserAgent))
	return err
}
