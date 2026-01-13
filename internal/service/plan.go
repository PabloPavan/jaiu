package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type PlanService struct {
	repo ports.PlanRepository
	now  func() time.Time
}

func NewPlanService(repo ports.PlanRepository) *PlanService {
	return &PlanService{repo: repo, now: time.Now}
}

func (s *PlanService) Create(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	if plan.Active == false {
		plan.Active = true
	}

	now := s.now()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	return s.repo.Create(ctx, plan)
}

func (s *PlanService) Update(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	plan.UpdatedAt = s.now()
	return s.repo.Update(ctx, plan)
}

func (s *PlanService) Deactivate(ctx context.Context, planID string) (domain.Plan, error) {
	plan, err := s.repo.FindByID(ctx, planID)
	if err != nil {
		return domain.Plan{}, err
	}

	plan.Active = false
	plan.UpdatedAt = s.now()

	return s.repo.Update(ctx, plan)
}

func (s *PlanService) ListActive(ctx context.Context) ([]domain.Plan, error) {
	return s.repo.ListActive(ctx)
}
