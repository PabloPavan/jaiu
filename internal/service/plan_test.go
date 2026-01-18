package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa Create setando timestamps e persistindo no repositorio.
func TestPlanServiceCreate(t *testing.T) {
	repo := &planRepoFake{}
	service := NewPlanService(repo, nil, nil)
	now := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	plan, err := service.Create(context.Background(), domain.Plan{Name: "Plano", DurationDays: 30, PriceCents: 1000, Active: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.CreatedAt.Equal(now) || !plan.UpdatedAt.Equal(now) {
		t.Fatal("expected timestamps to be set")
	}
}

// Testa Update encerrando assinaturas quando plano desativado.
func TestPlanServiceUpdateEndsSubscriptions(t *testing.T) {
	repo := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", Active: true},
		},
	}
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", PlanID: "plan-1", Status: domain.SubscriptionActive},
		},
	}
	service := NewPlanService(repo, subRepo, nil)

	updated, err := service.Update(context.Background(), domain.Plan{ID: "plan-1", Active: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Active {
		t.Fatal("expected plan to be inactive")
	}
	if subRepo.subscriptions["sub-1"].Status != domain.SubscriptionEnded {
		t.Fatalf("expected subscription ended, got %q", subRepo.subscriptions["sub-1"].Status)
	}
}

// Testa Deactivate carregando plano e encerrando assinaturas.
func TestPlanServiceDeactivate(t *testing.T) {
	repo := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", Active: true},
		},
	}
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", PlanID: "plan-1", Status: domain.SubscriptionActive},
		},
	}
	service := NewPlanService(repo, subRepo, nil)

	updated, err := service.Deactivate(context.Background(), "plan-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Active {
		t.Fatal("expected plan to be inactive")
	}
	if subRepo.subscriptions["sub-1"].Status != domain.SubscriptionEnded {
		t.Fatalf("expected subscription ended, got %q", subRepo.subscriptions["sub-1"].Status)
	}
}
