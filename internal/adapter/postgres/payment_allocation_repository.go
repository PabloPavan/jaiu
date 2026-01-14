package postgres

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentAllocationRepository struct {
	queries *sqlc.Queries
}

func NewPaymentAllocationRepository(pool *pgxpool.Pool) *PaymentAllocationRepository {
	return &PaymentAllocationRepository{queries: sqlc.New(pool)}
}

func (r *PaymentAllocationRepository) Create(ctx context.Context, allocation domain.PaymentAllocation) error {
	paymentID, err := stringToUUID(allocation.PaymentID)
	if err != nil || !paymentID.Valid {
		return err
	}
	periodID, err := stringToUUID(allocation.BillingPeriodID)
	if err != nil || !periodID.Valid {
		return err
	}

	if allocation.AmountCents <= 0 {
		return errors.New("valor de alocacao invalido")
	}

	params := sqlc.CreatePaymentAllocationParams{
		PaymentID:       paymentID,
		BillingPeriodID: periodID,
		AmountCents:     allocation.AmountCents,
	}

	return r.queries.CreatePaymentAllocation(ctx, params)
}

func (r *PaymentAllocationRepository) ListByPayment(ctx context.Context, paymentID string) ([]domain.PaymentAllocation, error) {
	uuidValue, err := stringToUUID(paymentID)
	if err != nil || !uuidValue.Valid {
		return nil, err
	}

	allocations, err := r.queries.ListPaymentAllocationsByPayment(ctx, uuidValue)
	if err != nil {
		return nil, err
	}

	result := make([]domain.PaymentAllocation, 0, len(allocations))
	for _, allocation := range allocations {
		result = append(result, domain.PaymentAllocation{
			PaymentID:       uuidToString(allocation.PaymentID),
			BillingPeriodID: uuidToString(allocation.BillingPeriodID),
			AmountCents:     allocation.AmountCents,
			CreatedAt:       timeFrom(allocation.CreatedAt),
		})
	}

	return result, nil
}

func (r *PaymentAllocationRepository) DeleteByPayment(ctx context.Context, paymentID string) error {
	uuidValue, err := stringToUUID(paymentID)
	if err != nil || !uuidValue.Valid {
		return err
	}

	return r.queries.DeletePaymentAllocationsByPayment(ctx, uuidValue)
}
