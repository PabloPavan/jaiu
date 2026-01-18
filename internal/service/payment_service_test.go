package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa Register validando assinatura obrigatoria.
func TestPaymentServiceRegisterMissingSubscription(t *testing.T) {
	service := NewPaymentService(&paymentRepoFake{}, nil, nil, nil, nil, nil, nil, nil)

	if _, err := service.Register(context.Background(), domain.Payment{AmountCents: 100}); err == nil {
		t.Fatal("expected error for missing subscription")
	}
}

// Testa Register validando valor do pagamento.
func TestPaymentServiceRegisterMissingAmount(t *testing.T) {
	service := NewPaymentService(&paymentRepoFake{}, nil, nil, nil, nil, nil, nil, nil)

	if _, err := service.Register(context.Background(), domain.Payment{SubscriptionID: "sub-1"}); err == nil {
		t.Fatal("expected error for missing amount")
	}
}

// Testa Register falhando quando dependencias nao estao configuradas.
func TestPaymentServiceRegisterMissingDeps(t *testing.T) {
	service := NewPaymentService(&paymentRepoFake{}, nil, nil, nil, nil, nil, nil, nil)

	if _, err := service.Register(context.Background(), domain.Payment{SubscriptionID: "sub-1", AmountCents: 100}); err == nil {
		t.Fatal("expected error for missing dependencies")
	}
}

// Testa Register retornando pagamento existente via idempotencia.
func TestPaymentServiceRegisterIdempotent(t *testing.T) {
	existing := domain.Payment{ID: "payment-1"}
	payments := &paymentRepoFake{
		payments:      map[string]domain.Payment{"payment-1": existing},
		byIdempotency: map[string]string{"idem": "payment-1"},
	}
	service := NewPaymentService(payments, &subscriptionRepoFake{}, &planRepoFake{}, &billingPeriodRepoFake{}, &balanceRepoFake{}, &paymentAllocationRepoFake{}, nil, nil)

	payment, err := service.Register(context.Background(), domain.Payment{
		SubscriptionID: "sub-1",
		AmountCents:    100,
		Method:         domain.PaymentCash,
		IdempotencyKey: "idem",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.ID != "payment-1" {
		t.Fatalf("expected existing payment, got %q", payment.ID)
	}
}

// Testa Register bloqueando metodo invalido.
func TestPaymentServiceRegisterInvalidMethod(t *testing.T) {
	payments := &paymentRepoFake{}
	subscriptions := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {
				ID:         "sub-1",
				PlanID:     "plan-1",
				Status:     domain.SubscriptionActive,
				StartDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PaymentDay: 1,
			},
		},
	}
	plans := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", DurationDays: 30, PriceCents: 1000},
		},
	}
	periods := &billingPeriodRepoFake{}
	allocations := &paymentAllocationRepoFake{}
	balances := &balanceRepoFake{balances: map[string]domain.SubscriptionBalance{"sub-1": {SubscriptionID: "sub-1"}}}

	service := NewPaymentService(payments, subscriptions, plans, periods, balances, allocations, nil, nil)
	service.now = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

	if _, err := service.Register(context.Background(), domain.Payment{
		SubscriptionID: "sub-1",
		AmountCents:    100,
		Method:         domain.PaymentMethod("x"),
	}); err == nil {
		t.Fatal("expected error for invalid method")
	}
}

// Testa Update bloqueando alteracao de valor.
func TestPaymentServiceUpdateBlocksAmountChange(t *testing.T) {
	current := domain.Payment{
		ID:             "payment-1",
		SubscriptionID: "sub-1",
		AmountCents:    100,
		Method:         domain.PaymentCash,
		Status:         domain.PaymentConfirmed,
		PaidAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Kind:           domain.PaymentFull,
	}
	repo := &paymentRepoFake{payments: map[string]domain.Payment{"payment-1": current}}
	service := NewPaymentService(repo, nil, nil, nil, nil, nil, nil, nil)

	if _, err := service.Update(context.Background(), domain.Payment{ID: "payment-1", AmountCents: 200}); err == nil {
		t.Fatal("expected error when changing amount")
	}
}

// Testa Update permitindo alteracoes seguras.
func TestPaymentServiceUpdateSuccess(t *testing.T) {
	current := domain.Payment{
		ID:             "payment-1",
		SubscriptionID: "sub-1",
		AmountCents:    100,
		Method:         domain.PaymentCash,
		Status:         domain.PaymentConfirmed,
		PaidAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Kind:           domain.PaymentFull,
	}
	repo := &paymentRepoFake{payments: map[string]domain.Payment{"payment-1": current}}
	service := NewPaymentService(repo, nil, nil, nil, nil, nil, nil, nil)

	updated, err := service.Update(context.Background(), domain.Payment{ID: "payment-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != "payment-1" {
		t.Fatalf("expected payment-1, got %q", updated.ID)
	}
}

// Testa Reverse retornando imediatamente quando ja estornado.
func TestPaymentServiceReverseAlreadyReversed(t *testing.T) {
	repo := &paymentRepoFake{
		payments: map[string]domain.Payment{
			"payment-1": {ID: "payment-1", Status: domain.PaymentReversed},
		},
	}
	service := NewPaymentService(repo, nil, nil, nil, nil, nil, nil, nil)

	payment, err := service.Reverse(context.Background(), "payment-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != domain.PaymentReversed {
		t.Fatalf("expected reversed status, got %q", payment.Status)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("expected no updates, got %d", repo.updateCalls)
	}
}

// Testa Reverse exigindo dependencias completas quando ha alocacoes.
func TestPaymentServiceReverseMissingSubscriptions(t *testing.T) {
	repo := &paymentRepoFake{
		payments: map[string]domain.Payment{
			"payment-1": {ID: "payment-1", Status: domain.PaymentConfirmed},
		},
	}
	service := NewPaymentService(repo, nil, nil, &billingPeriodRepoFake{}, nil, &paymentAllocationRepoFake{}, nil, nil)

	if _, err := service.Reverse(context.Background(), "payment-1"); err == nil {
		t.Fatal("expected error when subscriptions are missing")
	}
}
