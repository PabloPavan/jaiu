package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	queries *sqlc.Queries
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{queries: sqlc.New(pool)}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	studentID, err := stringToUUID(subscription.StudentID)
	if err != nil || !studentID.Valid {
		return domain.Subscription{}, err
	}
	planID, err := stringToUUID(subscription.PlanID)
	if err != nil || !planID.Valid {
		return domain.Subscription{}, err
	}

	params := sqlc.CreateSubscriptionParams{
		StudentID:  studentID,
		PlanID:     planID,
		StartDate:  dateTo(&subscription.StartDate),
		EndDate:    dateTo(&subscription.EndDate),
		Status:     string(subscription.Status),
		PriceCents: subscription.PriceCents,
		PaymentDay: int32(subscription.PaymentDay),
		AutoRenew:  subscription.AutoRenew,
	}

	created, err := r.queries.CreateSubscription(ctx, params)
	if err != nil {
		return domain.Subscription{}, err
	}

	return mapSubscription(created), nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	id, err := stringToUUID(subscription.ID)
	if err != nil || !id.Valid {
		return domain.Subscription{}, err
	}

	params := sqlc.UpdateSubscriptionParams{
		ID:         id,
		StartDate:  dateTo(&subscription.StartDate),
		EndDate:    dateTo(&subscription.EndDate),
		Status:     string(subscription.Status),
		PriceCents: subscription.PriceCents,
		PaymentDay: int32(subscription.PaymentDay),
		AutoRenew:  subscription.AutoRenew,
	}

	updated, err := r.queries.UpdateSubscription(ctx, params)
	if err != nil {
		return domain.Subscription{}, err
	}

	return mapSubscription(updated), nil
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (domain.Subscription, error) {
	uuidValue, err := stringToUUID(id)
	if err != nil || !uuidValue.Valid {
		return domain.Subscription{}, err
	}

	subscription, err := r.queries.GetSubscription(ctx, uuidValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, ports.ErrNotFound
		}
		return domain.Subscription{}, err
	}

	return mapSubscription(subscription), nil
}

func (r *SubscriptionRepository) ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error) {
	uuidValue, err := stringToUUID(studentID)
	if err != nil || !uuidValue.Valid {
		return nil, err
	}

	subscriptions, err := r.queries.ListSubscriptionsByStudent(ctx, uuidValue)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, mapSubscription(subscription))
	}

	return result, nil
}

func (r *SubscriptionRepository) ListDueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error) {
	params := sqlc.ListSubscriptionsDueBetweenParams{
		EndDate:   dateTo(&start),
		EndDate_2: dateTo(&end),
	}

	subscriptions, err := r.queries.ListSubscriptionsDueBetween(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, mapSubscription(subscription))
	}

	return result, nil
}

func (r *SubscriptionRepository) ListAutoRenew(ctx context.Context) ([]domain.Subscription, error) {
	subscriptions, err := r.queries.ListAutoRenewSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		result = append(result, mapSubscription(subscription))
	}

	return result, nil
}

func mapSubscription(subscription sqlc.Subscription) domain.Subscription {
	return domain.Subscription{
		ID:         uuidToString(subscription.ID),
		StudentID:  uuidToString(subscription.StudentID),
		PlanID:     uuidToString(subscription.PlanID),
		StartDate:  dateFromValue(subscription.StartDate),
		EndDate:    dateFromValue(subscription.EndDate),
		Status:     domain.SubscriptionStatus(subscription.Status),
		PriceCents: subscription.PriceCents,
		PaymentDay: int(subscription.PaymentDay),
		AutoRenew:  subscription.AutoRenew,
		CreatedAt:  timeFrom(subscription.CreatedAt),
		UpdatedAt:  timeFrom(subscription.UpdatedAt),
	}
}
