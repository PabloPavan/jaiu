package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Record(ctx context.Context, event domain.AuditEvent) error {
	if r == nil || r.pool == nil {
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
	_, err = r.pool.Exec(ctx, `
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
