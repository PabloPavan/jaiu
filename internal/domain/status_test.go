package domain

import "testing"

type validatable interface {
	IsValid() bool
}

// Testa IsValid para todos os status/roles expostos.
func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value validatable
		want  bool
	}{
		{"student-active", StudentActive, true},
		{"student-invalid", StudentStatus("unknown"), false},
		{"subscription-active", SubscriptionActive, true},
		{"subscription-invalid", SubscriptionStatus("unknown"), false},
		{"payment-confirmed", PaymentConfirmed, true},
		{"payment-invalid", PaymentStatus("unknown"), false},
		{"payment-kind-full", PaymentFull, true},
		{"payment-kind-invalid", PaymentKind("unknown"), false},
		{"payment-method-card", PaymentCard, true},
		{"payment-method-invalid", PaymentMethod("unknown"), false},
		{"billing-open", BillingOpen, true},
		{"billing-invalid", BillingPeriodStatus("unknown"), false},
		{"role-admin", RoleAdmin, true},
		{"role-invalid", UserRole("unknown"), false},
	}

	for _, tt := range tests {
		if got := tt.value.IsValid(); got != tt.want {
			t.Fatalf("%s: expected %v, got %v", tt.name, tt.want, got)
		}
	}
}
