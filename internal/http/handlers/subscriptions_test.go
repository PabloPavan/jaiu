package handlers

import (
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa o parse do status de assinatura com default e invalidacao.
func TestParseSubscriptionStatus(t *testing.T) {
	if got, err := parseSubscriptionStatus(""); err != nil || got != domain.SubscriptionActive {
		t.Fatalf("expected default active, got %q err=%v", got, err)
	}
	if _, err := parseSubscriptionStatus("invalid"); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// Testa normalizacao do filtro de status de assinatura.
func TestNormalizeSubscriptionStatus(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"", "all"},
		{"all", "all"},
		{string(domain.SubscriptionEnded), string(domain.SubscriptionEnded)},
		{string(domain.SubscriptionCanceled), string(domain.SubscriptionCanceled)},
		{string(domain.SubscriptionSuspended), string(domain.SubscriptionSuspended)},
		{"unknown", string(domain.SubscriptionActive)},
	}

	for _, tt := range tests {
		if got := normalizeSubscriptionStatus(tt.value); got != tt.want {
			t.Fatalf("normalizeSubscriptionStatus(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

// Testa a apresentacao de status em label e classe CSS.
func TestSubscriptionStatusPresentation(t *testing.T) {
	label, _ := subscriptionStatusPresentation(domain.SubscriptionEnded)
	if label != "Encerrada" {
		t.Fatalf("expected label Encerrada, got %q", label)
	}
	label, _ = subscriptionStatusPresentation(domain.SubscriptionSuspended)
	if label != "Suspensa" {
		t.Fatalf("expected label Suspensa, got %q", label)
	}
}

// Testa parse opcional de preco em centavos.
func TestParsePriceCentsOptional(t *testing.T) {
	if got, err := parsePriceCentsOptional(""); err != nil || got != 0 {
		t.Fatalf("expected empty to return 0 nil, got %d err=%v", got, err)
	}
	if got, err := parsePriceCentsOptional("1,50"); err != nil || got != 150 {
		t.Fatalf("expected 150, got %d err=%v", got, err)
	}
}

// Testa formatacao de data para exibicao.
func TestFormatDateBRValue(t *testing.T) {
	if got := formatDateBRValue(time.Time{}); got != "" {
		t.Fatalf("expected empty string for zero time, got %q", got)
	}
	date := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if got := formatDateBRValue(date); got != "02/01/2024" {
		t.Fatalf("expected 02/01/2024, got %q", got)
	}
}

// Testa o parse e validacao do dia de pagamento.
func TestParsePaymentDay(t *testing.T) {
	if _, err := parsePaymentDay(""); err == nil {
		t.Fatal("expected error for empty payment day")
	}
	if _, err := parsePaymentDay("0"); err == nil {
		t.Fatal("expected error for invalid payment day")
	}
	if got, err := parsePaymentDay("15"); err != nil || got != 15 {
		t.Fatalf("expected 15, got %d err=%v", got, err)
	}
}

// Testa conversao de inteiro para string com zero vazio.
func TestFormatInt(t *testing.T) {
	if got := formatInt(0); got != "" {
		t.Fatalf("expected empty string for zero, got %q", got)
	}
	if got := formatInt(12); got != "12" {
		t.Fatalf("expected 12, got %q", got)
	}
}
