package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa que endSubscriptions ignora repositorio nulo.
func TestEndSubscriptionsNilRepo(t *testing.T) {
	if err := endSubscriptions(context.Background(), nil, nil, time.Now()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// Testa que endSubscriptions encerra assinaturas ativas e ignora encerradas.
func TestEndSubscriptionsUpdatesStatus(t *testing.T) {
	repo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", Status: domain.SubscriptionActive},
			"sub-2": {ID: "sub-2", Status: domain.SubscriptionEnded},
		},
	}
	subscriptions := []domain.Subscription{
		repo.subscriptions["sub-1"],
		repo.subscriptions["sub-2"],
	}
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := endSubscriptions(context.Background(), repo, subscriptions, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.subscriptions["sub-1"].Status != domain.SubscriptionEnded {
		t.Fatalf("expected sub-1 to be ended, got %q", repo.subscriptions["sub-1"].Status)
	}
	if repo.subscriptions["sub-2"].Status != domain.SubscriptionEnded {
		t.Fatalf("expected sub-2 to stay ended, got %q", repo.subscriptions["sub-2"].Status)
	}
}
