package postgres

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionBalanceRepository struct {
	queries *sqlc.Queries
}

func NewSubscriptionBalanceRepository(pool *pgxpool.Pool) *SubscriptionBalanceRepository {
	return &SubscriptionBalanceRepository{queries: sqlc.New(pool)}
}

func (r *SubscriptionBalanceRepository) Get(ctx context.Context, subscriptionID string) (domain.SubscriptionBalance, error) {
	uuidValue, err := stringToUUID(subscriptionID)
	if err != nil || !uuidValue.Valid {
		return domain.SubscriptionBalance{}, err
	}

	balance, err := r.queries.GetSubscriptionBalance(ctx, uuidValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SubscriptionBalance{SubscriptionID: subscriptionID, CreditCents: 0}, nil
		}
		return domain.SubscriptionBalance{}, err
	}

	return mapSubscriptionBalance(balance), nil
}

func (r *SubscriptionBalanceRepository) Set(ctx context.Context, balance domain.SubscriptionBalance) (domain.SubscriptionBalance, error) {
	uuidValue, err := stringToUUID(balance.SubscriptionID)
	if err != nil || !uuidValue.Valid {
		return domain.SubscriptionBalance{}, err
	}

	params := sqlc.UpsertSubscriptionBalanceParams{
		SubscriptionID: uuidValue,
		CreditCents:    balance.CreditCents,
	}

	updated, err := r.queries.UpsertSubscriptionBalance(ctx, params)
	if err != nil {
		return domain.SubscriptionBalance{}, err
	}
	return mapSubscriptionBalance(updated), nil
}

func (r *SubscriptionBalanceRepository) Add(ctx context.Context, subscriptionID string, delta int64) (domain.SubscriptionBalance, error) {
	uuidValue, err := stringToUUID(subscriptionID)
	if err != nil || !uuidValue.Valid {
		return domain.SubscriptionBalance{}, err
	}

	params := sqlc.AddSubscriptionBalanceParams{
		SubscriptionID: uuidValue,
		CreditCents:    delta,
	}

	updated, err := r.queries.AddSubscriptionBalance(ctx, params)
	if err != nil {
		return domain.SubscriptionBalance{}, err
	}
	return mapSubscriptionBalance(updated), nil
}

func mapSubscriptionBalance(balance sqlc.SubscriptionBalance) domain.SubscriptionBalance {
	if !balance.SubscriptionID.Valid {
		return domain.SubscriptionBalance{}
	}
	return domain.SubscriptionBalance{
		SubscriptionID: uuidToString(balance.SubscriptionID),
		CreditCents:    balance.CreditCents,
		UpdatedAt:      timeFrom(balance.UpdatedAt),
	}
}
