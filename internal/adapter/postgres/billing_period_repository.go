package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingPeriodRepository struct {
	queries *sqlc.Queries
}

func NewBillingPeriodRepository(pool *pgxpool.Pool) *BillingPeriodRepository {
	return &BillingPeriodRepository{queries: sqlc.New(pool)}
}

func NewBillingPeriodRepositoryWithQueries(queries *sqlc.Queries) *BillingPeriodRepository {
	return &BillingPeriodRepository{queries: queries}
}

func (r *BillingPeriodRepository) Create(ctx context.Context, period domain.BillingPeriod) (domain.BillingPeriod, error) {
	subscriptionID, err := stringToUUID(period.SubscriptionID)
	if err != nil || !subscriptionID.Valid {
		return domain.BillingPeriod{}, err
	}

	params := sqlc.CreateBillingPeriodParams{
		SubscriptionID:  subscriptionID,
		PeriodStart:     dateTo(&period.PeriodStart),
		PeriodEnd:       dateTo(&period.PeriodEnd),
		AmountDueCents:  period.AmountDueCents,
		AmountPaidCents: period.AmountPaidCents,
		Status:          string(period.Status),
	}

	created, err := r.queries.CreateBillingPeriod(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "billing_periods_subscription_period_start_idx" {
			return domain.BillingPeriod{}, ports.ErrConflict
		}
		return domain.BillingPeriod{}, err
	}

	return mapBillingPeriod(created), nil
}

func (r *BillingPeriodRepository) Update(ctx context.Context, period domain.BillingPeriod) (domain.BillingPeriod, error) {
	id, err := stringToUUID(period.ID)
	if err != nil || !id.Valid {
		return domain.BillingPeriod{}, err
	}

	params := sqlc.UpdateBillingPeriodParams{
		ID:              id,
		AmountPaidCents: period.AmountPaidCents,
		Status:          string(period.Status),
	}

	updated, err := r.queries.UpdateBillingPeriod(ctx, params)
	if err != nil {
		return domain.BillingPeriod{}, err
	}

	return mapBillingPeriod(updated), nil
}

func (r *BillingPeriodRepository) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.BillingPeriod, error) {
	uuidValue, err := stringToUUID(subscriptionID)
	if err != nil || !uuidValue.Valid {
		return nil, err
	}

	periods, err := r.queries.ListBillingPeriodsBySubscription(ctx, uuidValue)
	if err != nil {
		return nil, err
	}

	result := make([]domain.BillingPeriod, 0, len(periods))
	for _, period := range periods {
		result = append(result, mapBillingPeriod(period))
	}

	return result, nil
}

func (r *BillingPeriodRepository) ListOpenBySubscription(ctx context.Context, subscriptionID string) ([]domain.BillingPeriod, error) {
	uuidValue, err := stringToUUID(subscriptionID)
	if err != nil || !uuidValue.Valid {
		return nil, err
	}

	periods, err := r.queries.ListOpenBillingPeriodsBySubscription(ctx, uuidValue)
	if err != nil {
		return nil, err
	}

	result := make([]domain.BillingPeriod, 0, len(periods))
	for _, period := range periods {
		result = append(result, mapBillingPeriod(period))
	}

	return result, nil
}

func (r *BillingPeriodRepository) MarkOverdue(ctx context.Context, subscriptionID string, now time.Time) error {
	uuidValue, err := stringToUUID(subscriptionID)
	if err != nil || !uuidValue.Valid {
		return err
	}

	params := sqlc.MarkBillingPeriodsOverdueParams{
		SubscriptionID: uuidValue,
		PeriodEnd:      dateTo(&now),
	}

	return r.queries.MarkBillingPeriodsOverdue(ctx, params)
}

func mapBillingPeriod(period sqlc.BillingPeriod) domain.BillingPeriod {
	if !period.ID.Valid {
		return domain.BillingPeriod{}
	}

	return domain.BillingPeriod{
		ID:              uuidToString(period.ID),
		SubscriptionID:  uuidToString(period.SubscriptionID),
		PeriodStart:     dateFromValue(period.PeriodStart),
		PeriodEnd:       dateFromValue(period.PeriodEnd),
		AmountDueCents:  period.AmountDueCents,
		AmountPaidCents: period.AmountPaidCents,
		Status:          domain.BillingPeriodStatus(period.Status),
		CreatedAt:       timeFrom(period.CreatedAt),
		UpdatedAt:       timeFrom(period.UpdatedAt),
	}
}
