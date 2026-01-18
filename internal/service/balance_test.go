package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa que applySubscriptionBalance retorna sem erro quando dependencias sao nulas.
func TestApplySubscriptionBalanceNilDeps(t *testing.T) {
	subscription := domain.Subscription{ID: "sub-1"}
	if err := applySubscriptionBalance(context.Background(), nil, nil, subscription, time.Now()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// Testa aplicacao de credito nos periodos abertos e atualizacao do saldo.
func TestApplySubscriptionBalanceAppliesCredit(t *testing.T) {
	balances := &balanceRepoFake{
		balances: map[string]domain.SubscriptionBalance{
			"sub-1": {SubscriptionID: "sub-1", CreditCents: 500},
		},
	}
	periods := &billingPeriodRepoFake{
		periods: map[string]domain.BillingPeriod{
			"p1": {
				ID:             "p1",
				SubscriptionID: "sub-1",
				PeriodStart:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				AmountDueCents: 300,
				Status:         domain.BillingOpen,
			},
			"p2": {
				ID:             "p2",
				SubscriptionID: "sub-1",
				PeriodStart:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				AmountDueCents: 300,
				Status:         domain.BillingOpen,
			},
		},
	}
	subscription := domain.Subscription{
		ID:         "sub-1",
		StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PaymentDay: 1,
	}
	today := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	if err := applySubscriptionBalance(context.Background(), balances, periods, subscription, today); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balances.balances["sub-1"].CreditCents != 0 {
		t.Fatalf("expected balance to be zero, got %d", balances.balances["sub-1"].CreditCents)
	}
	if periods.periods["p1"].AmountPaidCents != 300 {
		t.Fatalf("expected p1 paid 300, got %d", periods.periods["p1"].AmountPaidCents)
	}
	if periods.periods["p2"].AmountPaidCents != 200 {
		t.Fatalf("expected p2 paid 200, got %d", periods.periods["p2"].AmountPaidCents)
	}
}
