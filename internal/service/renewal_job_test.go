package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

// Testa Run retornando erro quando dependencias nao estao completas.
func TestRenewalJobRunMissingDeps(t *testing.T) {
	job := NewRenewalJob(nil, nil, nil, nil, nil)
	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected error for missing dependencies")
	}
}

// Testa Run processando assinaturas com renovacao automatica.
func TestRenewalJobRunProcessesSubscriptions(t *testing.T) {
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {
				ID:        "sub-1",
				PlanID:    "plan-1",
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				AutoRenew: true,
			},
		},
	}
	planRepo := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", DurationDays: 30, PriceCents: 1000},
		},
	}
	periodRepo := &billingPeriodRepoFake{}
	balanceRepo := &balanceRepoFake{
		balances: map[string]domain.SubscriptionBalance{
			"sub-1": {SubscriptionID: "sub-1", CreditCents: 0},
		},
	}
	job := NewRenewalJob(subRepo, planRepo, periodRepo, balanceRepo, nil)
	job.now = func() time.Time { return time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC) }

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(periodRepo.periods) == 0 {
		t.Fatal("expected billing periods to be created")
	}
}

// Testa Run propagando erro ao listar assinaturas com auto renew.
func TestRenewalJobRunListAutoRenewError(t *testing.T) {
	subRepo := &subscriptionRepoFake{listAutoErr: errors.New("boom")}
	job := NewRenewalJob(subRepo, &planRepoFake{}, &billingPeriodRepoFake{}, &balanceRepoFake{}, nil)

	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected error from list auto renew")
	}
}

// Testa Run propagando erro ao carregar o plano.
func TestRenewalJobRunPlanError(t *testing.T) {
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", PlanID: "plan-1", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AutoRenew: true},
		},
	}
	planRepo := &planRepoFake{findErr: errors.New("missing plan")}
	job := NewRenewalJob(subRepo, planRepo, &billingPeriodRepoFake{}, &balanceRepoFake{}, nil)

	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected error from plan lookup")
	}
}

type txRunnerSpy struct {
	called bool
	deps   ports.PaymentDependencies
	err    error
}

func (t *txRunnerSpy) RunSerializable(ctx context.Context, fn func(context.Context, ports.PaymentDependencies) error) error {
	t.called = true
	if t.err != nil {
		return t.err
	}
	return fn(ctx, t.deps)
}

// Testa Run usando o txRunner para processar a assinatura.
func TestRenewalJobRunUsesTxRunner(t *testing.T) {
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {
				ID:        "sub-1",
				PlanID:    "plan-1",
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				AutoRenew: true,
			},
		},
	}
	// Base repos existem apenas para passar na validacao inicial.
	basePlan := &planRepoFake{findErr: errors.New("base should not be used")}
	basePeriods := &billingPeriodRepoFake{createErr: errors.New("base should not be used")}
	baseBalances := &balanceRepoFake{getErr: errors.New("base should not be used")}

	txPlan := &planRepoFake{
		plans: map[string]domain.Plan{
			"plan-1": {ID: "plan-1", DurationDays: 30, PriceCents: 1000},
		},
	}
	txPeriods := &billingPeriodRepoFake{}
	txBalances := &balanceRepoFake{
		balances: map[string]domain.SubscriptionBalance{
			"sub-1": {SubscriptionID: "sub-1", CreditCents: 0},
		},
	}
	txRunner := &txRunnerSpy{
		deps: ports.PaymentDependencies{
			Plans:          txPlan,
			BillingPeriods: txPeriods,
			Balances:       txBalances,
		},
	}

	job := NewRenewalJob(subRepo, basePlan, basePeriods, baseBalances, txRunner)
	job.now = func() time.Time { return time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC) }

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !txRunner.called {
		t.Fatal("expected tx runner to be called")
	}
	if len(txPeriods.periods) == 0 {
		t.Fatal("expected periods created using tx deps")
	}
}

// Testa Run propagando erro do txRunner.
func TestRenewalJobRunTxRunnerError(t *testing.T) {
	subRepo := &subscriptionRepoFake{
		subscriptions: map[string]domain.Subscription{
			"sub-1": {ID: "sub-1", PlanID: "plan-1", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), AutoRenew: true},
		},
	}
	txRunner := &txRunnerSpy{err: errors.New("tx failed")}
	job := NewRenewalJob(subRepo, &planRepoFake{}, &billingPeriodRepoFake{}, &balanceRepoFake{}, txRunner)

	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected error from tx runner")
	}
}
