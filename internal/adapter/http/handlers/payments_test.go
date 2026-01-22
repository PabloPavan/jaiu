package handlers

import (
	"encoding/hex"
	"testing"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa o parse do metodo de pagamento com validacao.
func TestParsePaymentMethod(t *testing.T) {
	if _, err := parsePaymentMethod(""); err == nil {
		t.Fatal("expected error for empty method")
	}
	if _, err := parsePaymentMethod("invalid"); err == nil {
		t.Fatal("expected error for invalid method")
	}
	if got, err := parsePaymentMethod(string(domain.PaymentPix)); err != nil || got != domain.PaymentPix {
		t.Fatalf("expected pix, got %q err=%v", got, err)
	}
}

// Testa a geracao de idempotency key em formato hex.
func TestNewIdempotencyKey(t *testing.T) {
	key := newIdempotencyKey()
	if len(key) != 32 {
		t.Fatalf("expected 32 hex chars, got %d", len(key))
	}
	if _, err := hex.DecodeString(key); err != nil {
		t.Fatalf("expected hex string, got error: %v", err)
	}
}

// Testa o parse de status do pagamento com default.
func TestParsePaymentStatus(t *testing.T) {
	if got, err := parsePaymentStatus(""); err != nil || got != domain.PaymentConfirmed {
		t.Fatalf("expected default confirmed, got %q err=%v", got, err)
	}
	if _, err := parsePaymentStatus("invalid"); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// Testa a normalizacao de filtro de status.
func TestNormalizePaymentStatus(t *testing.T) {
	if got := normalizePaymentStatus(""); got != "all" {
		t.Fatalf("expected all, got %q", got)
	}
	if got := normalizePaymentStatus(string(domain.PaymentReversed)); got != string(domain.PaymentReversed) {
		t.Fatalf("expected reversed, got %q", got)
	}
}

// Testa apresentacao do status de pagamento.
func TestPaymentStatusPresentation(t *testing.T) {
	label, _ := paymentStatusPresentation(domain.PaymentReversed)
	if label != "Estornado" {
		t.Fatalf("expected Estornado, got %q", label)
	}
	label, _ = paymentStatusPresentation(domain.PaymentConfirmed)
	if label != "Confirmado" {
		t.Fatalf("expected Confirmado, got %q", label)
	}
}

// Testa apresentacao do tipo de pagamento.
func TestPaymentKindPresentation(t *testing.T) {
	label, _ := paymentKindPresentation(domain.PaymentAdvance)
	if label != "Adiantado" {
		t.Fatalf("expected Adiantado, got %q", label)
	}
	label, _ = paymentKindPresentation(domain.PaymentCredit)
	if label != "Credito" {
		t.Fatalf("expected Credito, got %q", label)
	}
}

// Testa label amigavel do metodo de pagamento.
func TestPaymentMethodLabel(t *testing.T) {
	if got := paymentMethodLabel(domain.PaymentCard); got != "Cartao" {
		t.Fatalf("expected Cartao, got %q", got)
	}
	if got := paymentMethodLabel(domain.PaymentCash); got != "Dinheiro" {
		t.Fatalf("expected Dinheiro, got %q", got)
	}
}

// Testa formatacao de valor para input de pagamento.
func TestFormatAmountInput(t *testing.T) {
	if got := formatAmountInput(250); got != "2,50" {
		t.Fatalf("expected 2,50, got %q", got)
	}
}
