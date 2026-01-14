package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type SubscriptionService struct {
	repo     ports.SubscriptionRepository
	plans    ports.PlanRepository
	students ports.StudentRepository
	now      func() time.Time
}

func NewSubscriptionService(repo ports.SubscriptionRepository, plans ports.PlanRepository, students ports.StudentRepository) *SubscriptionService {
	return &SubscriptionService{
		repo:     repo,
		plans:    plans,
		students: students,
		now:      time.Now,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	if subscription.StudentID == "" {
		return domain.Subscription{}, errors.New("student_id is required")
	}
	if subscription.PlanID == "" {
		return domain.Subscription{}, errors.New("plan_id is required")
	}

	if s.students != nil {
		if _, err := s.students.FindByID(ctx, subscription.StudentID); err != nil {
			return domain.Subscription{}, err
		}
	}

	var plan domain.Plan
	if s.plans != nil {
		loaded, err := s.plans.FindByID(ctx, subscription.PlanID)
		if err != nil {
			return domain.Subscription{}, err
		}
		plan = loaded
	}

	if subscription.Status == "" {
		subscription.Status = domain.SubscriptionActive
	}

	if subscription.StartDate.IsZero() {
		subscription.StartDate = dateOnly(s.now())
	}

	if subscription.EndDate.IsZero() && plan.DurationDays > 0 {
		subscription.EndDate = subscription.StartDate.AddDate(0, 0, plan.DurationDays)
	}

	if subscription.EndDate.IsZero() {
		return domain.Subscription{}, errors.New("end_date is required")
	}

	if subscription.EndDate.Before(subscription.StartDate) {
		return domain.Subscription{}, errors.New("end_date must be after start_date")
	}

	if subscription.PriceCents <= 0 && plan.PriceCents > 0 {
		subscription.PriceCents = plan.PriceCents
	}

	now := s.now()
	subscription.CreatedAt = now
	subscription.UpdatedAt = now

	return s.repo.Create(ctx, subscription)
}

func (s *SubscriptionService) Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	if subscription.StartDate.IsZero() {
		return domain.Subscription{}, errors.New("start_date is required")
	}
	if subscription.Status == "" {
		subscription.Status = domain.SubscriptionActive
	}

	var plan domain.Plan
	if subscription.EndDate.IsZero() || subscription.PriceCents <= 0 {
		if s.plans == nil {
			return domain.Subscription{}, errors.New("plan repository not configured")
		}
		loaded, err := s.plans.FindByID(ctx, subscription.PlanID)
		if err != nil {
			return domain.Subscription{}, err
		}
		plan = loaded
	}

	if subscription.EndDate.IsZero() && plan.DurationDays > 0 {
		subscription.EndDate = subscription.StartDate.AddDate(0, 0, plan.DurationDays)
	}

	if subscription.EndDate.IsZero() {
		return domain.Subscription{}, errors.New("end_date is required")
	}

	if subscription.EndDate.Before(subscription.StartDate) {
		return domain.Subscription{}, errors.New("end_date must be after start_date")
	}

	if subscription.PriceCents <= 0 && plan.PriceCents > 0 {
		subscription.PriceCents = plan.PriceCents
	}

	subscription.UpdatedAt = s.now()
	return s.repo.Update(ctx, subscription)
}

func (s *SubscriptionService) FindByID(ctx context.Context, subscriptionID string) (domain.Subscription, error) {
	return s.repo.FindByID(ctx, subscriptionID)
}

func (s *SubscriptionService) Cancel(ctx context.Context, subscriptionID string) (domain.Subscription, error) {
	subscription, err := s.repo.FindByID(ctx, subscriptionID)
	if err != nil {
		return domain.Subscription{}, err
	}

	subscription.Status = domain.SubscriptionCanceled
	subscription.UpdatedAt = s.now()

	return s.repo.Update(ctx, subscription)
}

func (s *SubscriptionService) ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error) {
	return s.repo.ListByStudent(ctx, studentID)
}

func (s *SubscriptionService) DueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error) {
	return s.repo.ListDueBetween(ctx, start, end)
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
