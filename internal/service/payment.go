package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type PaymentService struct {
	repo          ports.PaymentRepository
	subscriptions ports.SubscriptionRepository
	now           func() time.Time
}

func NewPaymentService(repo ports.PaymentRepository, subscriptions ports.SubscriptionRepository) *PaymentService {
	return &PaymentService{repo: repo, subscriptions: subscriptions, now: time.Now}
}

func (s *PaymentService) Register(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.SubscriptionID == "" {
		return domain.Payment{}, errors.New("assinatura e obrigatoria")
	}
	if s.subscriptions != nil {
		if _, err := s.subscriptions.FindByID(ctx, payment.SubscriptionID); err != nil {
			return domain.Payment{}, err
		}
	}
	if payment.Status == "" {
		payment.Status = domain.PaymentConfirmed
	}

	if payment.PaidAt.IsZero() {
		payment.PaidAt = s.now()
	}

	if payment.AmountCents <= 0 {
		return domain.Payment{}, errors.New("valor deve ser maior que zero")
	}

	payment.CreatedAt = s.now()

	return s.repo.Create(ctx, payment)
}

func (s *PaymentService) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.SubscriptionID == "" {
		return domain.Payment{}, errors.New("assinatura e obrigatoria")
	}
	if s.subscriptions != nil {
		if _, err := s.subscriptions.FindByID(ctx, payment.SubscriptionID); err != nil {
			return domain.Payment{}, err
		}
	}
	if payment.PaidAt.IsZero() {
		payment.PaidAt = s.now()
	}
	if payment.Status == "" {
		payment.Status = domain.PaymentConfirmed
	}
	if payment.AmountCents <= 0 {
		return domain.Payment{}, errors.New("valor deve ser maior que zero")
	}

	return s.repo.Update(ctx, payment)
}

func (s *PaymentService) FindByID(ctx context.Context, paymentID string) (domain.Payment, error) {
	return s.repo.FindByID(ctx, paymentID)
}

func (s *PaymentService) Reverse(ctx context.Context, paymentID string) (domain.Payment, error) {
	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return domain.Payment{}, err
	}

	payment.Status = domain.PaymentReversed
	return s.repo.Update(ctx, payment)
}

func (s *PaymentService) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error) {
	return s.repo.ListBySubscription(ctx, subscriptionID)
}

func (s *PaymentService) ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error) {
	return s.repo.ListByPeriod(ctx, start, end)
}
