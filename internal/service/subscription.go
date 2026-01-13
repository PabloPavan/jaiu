package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type SubscriptionService struct {
	repo ports.SubscriptionRepository
	now  func() time.Time
}

func NewSubscriptionService(repo ports.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo, now: time.Now}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	if subscription.Status == "" {
		subscription.Status = domain.SubscriptionActive
	}

	now := s.now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	return s.repo.Create(ctx, subscription)
}

func (s *SubscriptionService) Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	subscription.UpdatedAt = s.now()
	return s.repo.Update(ctx, subscription)
}

func (s *SubscriptionService) ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error) {
	return s.repo.ListByStudent(ctx, studentID)
}

func (s *SubscriptionService) DueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error) {
	return s.repo.ListDueBetween(ctx, start, end)
}
