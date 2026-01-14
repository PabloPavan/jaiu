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
	queries *sqlc.Queries
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{queries: sqlc.New(pool)}
}

func (r *PaymentRepository) Create(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	subscriptionID, err := stringToUUID(payment.SubscriptionID)
	if err != nil || !subscriptionID.Valid {
		return domain.Payment{}, err
	}

	kind := string(payment.Kind)
	if kind == "" {
		kind = string(domain.PaymentFull)
	}

	params := sqlc.CreatePaymentParams{
		SubscriptionID: subscriptionID,
		PaidAt:         pgtype.Timestamptz{Time: payment.PaidAt, Valid: true},
		AmountCents:    payment.AmountCents,
		Method:         string(payment.Method),
		Reference:      textTo(payment.Reference),
		Notes:          textTo(payment.Notes),
		Status:         string(payment.Status),
		Kind:           kind,
		CreditCents:    payment.CreditCents,
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

	subscriptionID, err := stringToUUID(payment.SubscriptionID)
	if err != nil || !subscriptionID.Valid {
		return domain.Payment{}, err
	}

	kind := string(payment.Kind)
	if kind == "" {
		kind = string(domain.PaymentFull)
	}

	params := sqlc.UpdatePaymentParams{
		ID:             id,
		SubscriptionID: subscriptionID,
		PaidAt:         pgtype.Timestamptz{Time: payment.PaidAt, Valid: true},
		AmountCents:    payment.AmountCents,
		Method:         string(payment.Method),
		Reference:      textTo(payment.Reference),
		Notes:          textTo(payment.Notes),
		Status:         string(payment.Status),
		Kind:           kind,
		CreditCents:    payment.CreditCents,
	}

	updated, err := r.queries.UpdatePayment(ctx, params)
	if err != nil {
		return domain.Payment{}, err
	}

	return mapPayment(updated), nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (domain.Payment, error) {
	uuidValue, err := stringToUUID(id)
	if err != nil || !uuidValue.Valid {
		return domain.Payment{}, err
	}

	payment, err := r.queries.GetPayment(ctx, uuidValue)
	if err != nil {
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
	if !payment.ID.Valid {
		return domain.Payment{}
	}

	result := domain.Payment{
		ID:             uuidToString(payment.ID),
		SubscriptionID: uuidToString(payment.SubscriptionID),
		PaidAt:         timeFrom(payment.PaidAt),
		AmountCents:    payment.AmountCents,
		Method:         domain.PaymentMethod(payment.Method),
		Reference:      textFrom(payment.Reference),
		Notes:          textFrom(payment.Notes),
		Status:         domain.PaymentStatus(payment.Status),
		Kind:           domain.PaymentKind(payment.Kind),
		CreditCents:    payment.CreditCents,
		CreatedAt:      timeFrom(payment.CreatedAt),
	}
	if result.Kind == "" {
		result.Kind = domain.PaymentFull
	}

	return result
}
