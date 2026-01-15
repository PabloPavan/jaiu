package postgres

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanRepository struct {
	queries *sqlc.Queries
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{queries: sqlc.New(pool)}
}

func NewPlanRepositoryWithQueries(queries *sqlc.Queries) *PlanRepository {
	return &PlanRepository{queries: queries}
}

func (r *PlanRepository) Create(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	params := sqlc.CreatePlanParams{
		Name:         plan.Name,
		DurationDays: int32(plan.DurationDays),
		PriceCents:   plan.PriceCents,
		Active:       plan.Active,
		Description:  pgtype.Text{String: plan.Description, Valid: plan.Description != ""},
	}

	created, err := r.queries.CreatePlan(ctx, params)
	if err != nil {
		return domain.Plan{}, err
	}

	return mapPlan(created), nil
}

func (r *PlanRepository) Update(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	id, err := stringToUUID(plan.ID)
	if err != nil || !id.Valid {
		return domain.Plan{}, err
	}

	params := sqlc.UpdatePlanParams{
		ID:           id,
		Name:         plan.Name,
		DurationDays: int32(plan.DurationDays),
		PriceCents:   plan.PriceCents,
		Active:       plan.Active,
		Description:  pgtype.Text{String: plan.Description, Valid: plan.Description != ""},
	}

	updated, err := r.queries.UpdatePlan(ctx, params)
	if err != nil {
		return domain.Plan{}, err
	}

	return mapPlan(updated), nil
}

func (r *PlanRepository) FindByID(ctx context.Context, id string) (domain.Plan, error) {
	uuidValue, err := stringToUUID(id)
	if err != nil || !uuidValue.Valid {
		return domain.Plan{}, err
	}

	plan, err := r.queries.GetPlan(ctx, uuidValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Plan{}, ports.ErrNotFound
		}
		return domain.Plan{}, err
	}

	return mapPlan(plan), nil
}

func (r *PlanRepository) ListActive(ctx context.Context) ([]domain.Plan, error) {
	plans, err := r.queries.ListActivePlans(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Plan, 0, len(plans))
	for _, plan := range plans {
		result = append(result, mapPlan(plan))
	}

	return result, nil
}

func mapPlan(plan sqlc.Plan) domain.Plan {
	return domain.Plan{
		ID:           uuidToString(plan.ID),
		Name:         plan.Name,
		DurationDays: int(plan.DurationDays),
		PriceCents:   plan.PriceCents,
		Active:       plan.Active,
		Description:  textFrom(plan.Description),
		CreatedAt:    timeFrom(plan.CreatedAt),
		UpdatedAt:    timeFrom(plan.UpdatedAt),
	}
}
