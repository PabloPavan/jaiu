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
	audit         ports.AuditRepository
	now           func() time.Time
}

func NewPlanService(repo ports.PlanRepository, subscriptions ports.SubscriptionRepository, audit ports.AuditRepository) *PlanService {
	return &PlanService{repo: repo, subscriptions: subscriptions, audit: audit, now: time.Now}
}

func (s *PlanService) Create(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	now := s.now()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	metadata := map[string]any{
		"active":        plan.Active,
		"duration_days": plan.DurationDays,
		"price_cents":   plan.PriceCents,
	}
	recordAuditAttempt(ctx, s.audit, "plan.create", "plan", plan.ID, metadata)

	created, err := s.repo.Create(ctx, plan)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "plan.create", "plan", plan.ID, metadata, err)
		return domain.Plan{}, err
	}

	recordAuditSuccess(ctx, s.audit, "plan.create", "plan", created.ID, metadata)
	return created, nil
}

func (s *PlanService) Update(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	plan.UpdatedAt = s.now()
	metadata := map[string]any{
		"active":        plan.Active,
		"duration_days": plan.DurationDays,
		"price_cents":   plan.PriceCents,
	}
	recordAuditAttempt(ctx, s.audit, "plan.update", "plan", plan.ID, metadata)

	updated, err := s.repo.Update(ctx, plan)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "plan.update", "plan", plan.ID, metadata, err)
		return domain.Plan{}, err
	}
	if !updated.Active {
		if err := s.endSubscriptionsForPlan(ctx, updated.ID); err != nil {
			recordAuditFailure(ctx, s.audit, "plan.update", "plan", updated.ID, metadata, err)
			return updated, err
		}
	}
	recordAuditSuccess(ctx, s.audit, "plan.update", "plan", updated.ID, metadata)
	return updated, nil
}

func (s *PlanService) Deactivate(ctx context.Context, planID string) (domain.Plan, error) {
	recordAuditAttempt(ctx, s.audit, "plan.deactivate", "plan", planID, nil)

	plan, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "plan.deactivate", "plan", planID, nil, err)
		return domain.Plan{}, err
	}

	plan.Active = false
	plan.UpdatedAt = s.now()

	updated, err := s.repo.Update(ctx, plan)
	if err != nil {
		recordAuditFailure(ctx, s.audit, "plan.deactivate", "plan", planID, nil, err)
		return domain.Plan{}, err
	}
	if err := s.endSubscriptionsForPlan(ctx, updated.ID); err != nil {
		recordAuditFailure(ctx, s.audit, "plan.deactivate", "plan", updated.ID, nil, err)
		return updated, err
	}
	recordAuditSuccess(ctx, s.audit, "plan.deactivate", "plan", updated.ID, map[string]any{
		"active": updated.Active,
	})
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
