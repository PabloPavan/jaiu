package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa Create validando campos obrigatorios.
func TestSubscriptionServiceCreateValidation(t *testing.T) {
	service := NewSubscriptionService(&subscriptionRepoFake{}, nil, nil, nil)

	if _, err := service.Create(context.Background(), domain.Subscription{}); err == nil {
		t.Fatal("expected error for missing student and plan")
	}
	if _, err := service.Create(context.Background(), domain.Subscription{StudentID: "s1"}); err == nil {
		t.Fatal("expected error for missing plan")
	}
	if _, err := service.Create(context.Background(), domain.Subscription{StudentID: "s1", PlanID: "p1", Status: domain.SubscriptionStatus("x")}); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// Testa Create preenchendo defaults e datas.
func TestSubscriptionServiceCreateDefaults(t *testing.T) {
	subRepo := &subscriptionRepoFake{}
	planRepo := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", DurationDays: 30, PriceCents: 1000},
		},
	}
	service := NewSubscriptionService(subRepo, planRepo, nil, nil)
	now := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	subscription, err := service.Create(context.Background(), domain.Subscription{
		StudentID: "student-1",
		PlanID:    "plan-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subscription.Status != domain.SubscriptionActive {
		t.Fatalf("expected active status, got %q", subscription.Status)
	}
	if subscription.PaymentDay != subscription.StartDate.Day() {
		t.Fatalf("expected payment day to match start date")
	}
	if subscription.EndDate.IsZero() {
		t.Fatal("expected end date to be set")
	}
}

// Testa Update validando datas e status.
func TestSubscriptionServiceUpdateValidation(t *testing.T) {
	subRepo := &subscriptionRepoFake{}
	planRepo := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", DurationDays: 30, PriceCents: 1000},
		},
	}
	service := NewSubscriptionService(subRepo, planRepo, nil, nil)

	if _, err := service.Update(context.Background(), domain.Subscription{ID: "sub-1"}); err == nil {
		t.Fatal("expected error for missing start date")
	}
	start := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	if _, err := service.Update(context.Background(), domain.Subscription{ID: "sub-1", StartDate: start, PaymentDay: 40}); err == nil {
		t.Fatal("expected error for invalid payment day")
	}
	if _, err := service.Update(context.Background(), domain.Subscription{
		ID:         "sub-1",
		StartDate:  start,
		EndDate:    start.AddDate(0, 0, -1),
		PaymentDay: 1,
	}); err == nil {
		t.Fatal("expected error for end date before start date")
	}
}

// Testa Cancel mudando status e salvando no repositorio.
func TestSubscriptionServiceCancel(t *testing.T) {
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", Status: domain.SubscriptionActive},
		},
	}
	service := NewSubscriptionService(subRepo, nil, nil, nil)

	updated, err := service.Cancel(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != domain.SubscriptionCanceled {
		t.Fatalf("expected canceled, got %q", updated.Status)
	}
}
