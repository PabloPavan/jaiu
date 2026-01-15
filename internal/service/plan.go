package service

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type PlanService struct {
	repo          ports.PlanRepository
	subscriptions ports.SubscriptionRepository
	now           func() time.Time
}

func NewPlanService(repo ports.PlanRepository, subscriptions ports.SubscriptionRepository) *PlanService {
	return &PlanService{repo: repo, subscriptions: subscriptions, now: time.Now}
}

func (s *PlanService) Create(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	now := s.now()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	return s.repo.Create(ctx, plan)
}

func (s *PlanService) Update(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	plan.UpdatedAt = s.now()
	updated, err := s.repo.Update(ctx, plan)
	if err != nil {
		return domain.Plan{}, err
	}
	if !updated.Active {
		if err := s.endSubscriptionsForPlan(ctx, updated.ID); err != nil {
			return updated, err
		}
	}
	return updated, nil
}

func (s *PlanService) Deactivate(ctx context.Context, planID string) (domain.Plan, error) {
	plan, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return domain.Plan{}, err
	}

	plan.Active = false
	plan.UpdatedAt = s.now()

	updated, err := s.repo.Update(ctx, plan)
	if err != nil {
		return domain.Plan{}, err
	}
	if err := s.endSubscriptionsForPlan(ctx, updated.ID); err != nil {
		return updated, err
	}
	return updated, nil
}

func (s *PlanService) FindByID(ctx context.Context, planID string) (domain.Plan, error) {
	return s.repo.FindByID(ctx, planID)
}

func (s *PlanService) ListActive(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.ListActive(ctx)
}

func (s *PlanService) endSubscriptionsForPlan(ctx context.Context, planID string) error {
	if s.subscriptions == nil {
		return errors.New("assinaturas indisponiveis")
	}
	subscriptions, err := s.subscriptions.ListByPlan(ctx, planID)
	if err != nil {
		return err
	}
	return endSubscriptions(ctx, s.subscriptions, subscriptions, s.now())
}
