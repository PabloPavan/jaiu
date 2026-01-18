package service

import (
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa buildPeriod criando datas e status corretos.
func TestBuildPeriod(t *testing.T) {
	start := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	period := buildPeriod("sub-1", start, 30, 1000)
	if period.PeriodStart.Hour() != 0 || period.PeriodStart.Minute() != 0 {
		t.Fatal("expected start date normalized to midnight")
	}
	if period.AmountDueCents != 1000 || period.Status != domain.BillingOpen {
		t.Fatalf("unexpected period fields: %#v", period)
	}
}

// Testa resolucao de status do periodo conforme pagamento e vencimento.
func TestResolvePeriodStatus(t *testing.T) {
	period := domain.BillingPeriod{
		PeriodStart:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AmountDueCents: 1000,
	}
	today := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	period.AmountPaidCents = 1000
	if got := resolvePeriodStatus(period, today, 10); got != domain.BillingPaid {
		t.Fatalf("expected paid, got %q", got)
	}

	period.AmountPaidCents = 100
	if got := resolvePeriodStatus(period, today, 10); got != domain.BillingPartial {
		t.Fatalf("expected partial, got %q", got)
	}

	late := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	period.AmountPaidCents = 0
	if got := resolvePeriodStatus(period, late, 10); got != domain.BillingOverdue {
		t.Fatalf("expected overdue, got %q", got)
	}
}

// Testa escolha do tipo de pagamento com base nos flags.
func TestPaymentKind(t *testing.T) {
	if got := paymentKind(false, false, true); got != domain.PaymentCredit {
		t.Fatalf("expected credit, got %q", got)
	}
	if got := paymentKind(true, false, false); got != domain.PaymentPartial {
		t.Fatalf("expected partial, got %q", got)
	}
	if got := paymentKind(false, true, false); got != domain.PaymentAdvance {
		t.Fatalf("expected advance, got %q", got)
	}
	if got := paymentKind(false, false, false); got != domain.PaymentFull {
		t.Fatalf("expected full, got %q", got)
	}
}

// Testa utilitarios de minimo e comparacao de tempo.
func TestPaymentHelpers(t *testing.T) {
	if got := minInt64(2, 5); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !sameDayTime(now, now) {
		t.Fatal("expected same day times to be equal")
	}
}
