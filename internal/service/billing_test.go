package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa calculo de preco efetivo entre assinatura e plano.
func TestEffectivePriceCents(t *testing.T) {
	if _, err := effectivePriceCents(domain.Subscription{}, domain.Plan{}); err == nil {
		t.Fatal("expected error when both prices are zero")
	}
	if got, err := effectivePriceCents(domain.Subscription{PriceCents: 1500}, domain.Plan{PriceCents: 1000}); err != nil || got != 1500 {
		t.Fatalf("expected subscription price, got %d err=%v", got, err)
	}
	if got, err := effectivePriceCents(domain.Subscription{}, domain.Plan{PriceCents: 900}); err != nil || got != 900 {
		t.Fatalf("expected plan price, got %d err=%v", got, err)
	}
}

// Testa calculo do dia de pagamento efetivo.
func TestEffectivePaymentDay(t *testing.T) {
	if _, err := effectivePaymentDay(domain.Subscription{}); err == nil {
		t.Fatal("expected error for missing start date")
	}
	start := time.Date(2024, 4, 10, 12, 0, 0, 0, time.UTC)
	if got, err := effectivePaymentDay(domain.Subscription{StartDate: start}); err != nil || got != 10 {
		t.Fatalf("expected day 10, got %d err=%v", got, err)
	}
	if _, err := effectivePaymentDay(domain.Subscription{PaymentDay: 32, StartDate: start}); err == nil {
		t.Fatal("expected error for invalid payment day")
	}
}

// Testa ajuste do ultimo dia do mes quando o dia extrapola.
func TestClampPaymentDay(t *testing.T) {
	day := clampPaymentDay(31, 2024, time.April, time.UTC)
	if day != 30 {
		t.Fatalf("expected 30, got %d", day)
	}
}

// Testa calculo de vencimento quando o dia cai antes do inicio.
func TestDueDateForPeriod(t *testing.T) {
	start := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	due := dueDateForPeriod(start, 5)
	if due.Format("2006-01-02") != "2024-02-05" {
		t.Fatalf("expected 2024-02-05, got %s", due.Format("2006-01-02"))
	}
}

// Testa a chave de data gerada para comparacoes.
func TestDateKey(t *testing.T) {
	value := time.Date(2024, 5, 2, 15, 0, 0, 0, time.UTC)
	if got := dateKey(value); got != "2024-05-02" {
		t.Fatalf("expected 2024-05-02, got %q", got)
	}
}

// Testa criacao do primeiro periodo de cobranca.
func TestEnsureBillingPeriodsCreatesFirst(t *testing.T) {
	repo := &billingPeriodRepoFake{}
	subscription := domain.Subscription{
		ID:        "sub-1",
		StartDate: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
		AutoRenew: false,
	}
	plan := domain.Plan{DurationDays: 30, PriceCents: 1000}
	today := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	periods, err := ensureBillingPeriods(context.Background(), repo, subscription, plan, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(periods))
	}
	if periods[0].AmountDueCents != 1000 {
		t.Fatalf("expected amount due 1000, got %d", periods[0].AmountDueCents)
	}
}

// Testa criacao de renovacoes quando auto renew esta ativo.
func TestEnsureBillingPeriodsAutoRenew(t *testing.T) {
	repo := &billingPeriodRepoFake{
		periods: map[string]domain.BillingPeriod{
			"period-1": {
				ID:             "period-1",
				SubscriptionID: "sub-1",
				PeriodStart:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				AmountDueCents: 1000,
				Status:         domain.BillingOpen,
			},
		},
	}
	subscription := domain.Subscription{
		ID:        "sub-1",
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AutoRenew: true,
	}
	plan := domain.Plan{DurationDays: 30, PriceCents: 1000}
	today := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	periods, err := ensureBillingPeriods(context.Background(), repo, subscription, plan, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(periods) < 2 {
		t.Fatalf("expected at least 2 periods, got %d", len(periods))
	}
}

// Testa atualizacao de status quando periodo fica vencido.
func TestRefreshPeriodStatuses(t *testing.T) {
	repo := &billingPeriodRepoFake{
		periods: map[string]domain.BillingPeriod{
			"period-1": {
				ID:             "period-1",
				SubscriptionID: "sub-1",
				PeriodStart:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				AmountDueCents: 1000,
				Status:         domain.BillingOpen,
			},
		},
	}
	periods := []domain.BillingPeriod{repo.periods["period-1"]}
	today := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	updated, err := refreshPeriodStatuses(context.Background(), repo, periods, 1, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated[0].Status != domain.BillingOverdue {
		t.Fatalf("expected overdue, got %q", updated[0].Status)
	}
}
