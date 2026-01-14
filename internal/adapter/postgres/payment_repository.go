package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool, queries: sqlc.New(pool)}
}

func (r *PaymentRepository) Create(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	subscriptionID, err := stringToUUID(payment.SubscriptionID)
	if err != nil || !subscriptionID.Valid {
		return domain.Payment{}, err
	}

	params := sqlc.CreatePaymentParams{
		SubscriptionID: subscriptionID,
		PaidAt:         pgtype.Timestamptz{Time: payment.PaidAt, Valid: true},
		AmountCents:    payment.AmountCents,
		Method:         string(payment.Method),
		Reference:      textTo(payment.Reference),
		Notes:          textTo(payment.Notes),
		Status:         string(payment.Status),
	}

	created, err := r.queries.CreatePayment(ctx, params)
	if err != nil {
		return domain.Payment{}, err
	}

	return mapPayment(created), nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	id, err := stringToUUID(payment.ID)
	if err != nil || !id.Valid {
		return domain.Payment{}, err
	}

	query := `
		UPDATE payments
		SET paid_at = $2,
		    amount_cents = $3,
		    method = $4,
		    reference = $5,
		    notes = $6,
		    status = $7
		WHERE id = $1
		RETURNING id, subscription_id, paid_at, amount_cents, method, reference, notes, status, created_at
	`

	row := r.pool.QueryRow(ctx, query,
		id,
		pgtype.Timestamptz{Time: payment.PaidAt, Valid: true},
		payment.AmountCents,
		string(payment.Method),
		textTo(payment.Reference),
		textTo(payment.Notes),
		string(payment.Status),
	)

	var updated sqlc.Payment
	if err := row.Scan(
		&updated.ID,
		&updated.SubscriptionID,
		&updated.PaidAt,
		&updated.AmountCents,
		&updated.Method,
		&updated.Reference,
		&updated.Notes,
		&updated.Status,
		&updated.CreatedAt,
	); err != nil {
		return domain.Payment{}, err
	}

	return mapPayment(updated), nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (domain.Payment, error) {
	uuidValue, err := stringToUUID(id)
	if err != nil || !uuidValue.Valid {
		return domain.Payment{}, err
	}

	query := `
		SELECT id, subscription_id, paid_at, amount_cents, method, reference, notes, status, created_at
		FROM payments
		WHERE id = $1
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, uuidValue)

	var payment sqlc.Payment
	if err := row.Scan(
		&payment.ID,
		&payment.SubscriptionID,
		&payment.PaidAt,
		&payment.AmountCents,
		&payment.Method,
		&payment.Reference,
		&payment.Notes,
		&payment.Status,
		&payment.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Payment{}, ports.ErrNotFound
		}
		return domain.Payment{}, err
	}

	return mapPayment(payment), nil
}

func (r *PaymentRepository) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error) {
	uuidValue, err := stringToUUID(subscriptionID)
	if err != nil || !uuidValue.Valid {
		return nil, err
	}

	payments, err := r.queries.ListPaymentsBySubscription(ctx, uuidValue)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Payment, 0, len(payments))
	for _, payment := range payments {
		result = append(result, mapPayment(payment))
	}

	return result, nil
}

func (r *PaymentRepository) ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error) {
	params := sqlc.ListPaymentsByPeriodParams{
		PaidAt:   pgtype.Timestamptz{Time: start, Valid: true},
		PaidAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	}

	payments, err := r.queries.ListPaymentsByPeriod(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Payment, 0, len(payments))
	for _, payment := range payments {
		result = append(result, mapPayment(payment))
	}

	return result, nil
}

func mapPayment(payment sqlc.Payment) domain.Payment {
	return domain.Payment{
		ID:             uuidToString(payment.ID),
		SubscriptionID: uuidToString(payment.SubscriptionID),
		PaidAt:         timeFrom(payment.PaidAt),
		AmountCents:    payment.AmountCents,
		Method:         domain.PaymentMethod(payment.Method),
		Reference:      textFrom(payment.Reference),
		Notes:          textFrom(payment.Notes),
		Status:         domain.PaymentStatus(payment.Status),
		CreatedAt:      timeFrom(payment.CreatedAt),
	}
}
