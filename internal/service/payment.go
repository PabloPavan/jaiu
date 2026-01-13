package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type PaymentService struct {
	repo ports.PaymentRepository
	now  func() time.Time
}

func NewPaymentService(repo ports.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo, now: time.Now}
}

func (s *PaymentService) Register(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.Status == "" {
		payment.Status = domain.PaymentConfirmed
	}

	if payment.PaidAt.IsZero() {
		payment.PaidAt = s.now()
	}

	payment.CreatedAt = s.now()

	return s.repo.Create(ctx, payment)
}

func (s *PaymentService) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error) {
	return s.repo.ListBySubscription(ctx, subscriptionID)
}

func (s *PaymentService) ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error) {
	return s.repo.ListByPeriod(ctx, start, end)
}
