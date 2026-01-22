package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/view"
)

// Testa a formatacao de valores monetarios para exibicao.
func TestFormatCentsAndBRL(t *testing.T) {
	if got := formatCents(1234); got != "R$ 12,34" {
		t.Fatalf("expected formatCents to be R$ 12,34, got %q", got)
	}
	if got := formatBRL(500); got != "R$ 5,00" {
		t.Fatalf("expected formatBRL to be R$ 5,00, got %q", got)
	}
}

// Testa a formatacao de centavos para input de formulario.
func TestFormatCentsInput(t *testing.T) {
	if got := formatCentsInput(987); got != "9,87" {
		t.Fatalf("expected formatCentsInput to be 9,87, got %q", got)
	}
}

// Testa o parse de valores monetarios com virgula e ponto.
func TestParsePriceCents(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int64
	}{
		{"comma", "10,50", 1050},
		{"thousand", "1.234,56", 123456},
		{"dot", "12.34", 1234},
	}

	for _, tt := range tests {
		got, err := parsePriceCents(tt.value)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: expected %d, got %d", tt.name, tt.want, got)
		}
	}

	if _, err := parsePriceCents(" "); err == nil {
		t.Fatal("expected error for empty price")
	}
}

// Testa dados padrao do formulario de criacao de plano.
func TestPlanFormCreateData(t *testing.T) {
	data := planFormCreateData()
	if data.Title == "" || data.Action == "" || data.SubmitLabel == "" {
		t.Fatalf("expected base fields to be set, got %#v", data)
	}
	if !data.Active {
		t.Fatal("expected new plan to be active by default")
	}
}

// Testa dados do formulario de edicao de plano.
func TestPlanFormEditData(t *testing.T) {
	plan := domain.Plan{
		ID:           "plan-1",
		Name:         "Basico",
		DurationDays: 30,
		PriceCents:   1000,
		Description:  "desc",
		Active:       true,
	}
	data := planFormEditData(plan)
	if data.Action != "/plans/plan-1" {
		t.Fatalf("expected action to be /plans/plan-1, got %q", data.Action)
	}
	if !data.ShowDelete {
		t.Fatal("expected delete to be visible for existing plan")
	}
	if data.Price != "10,00" {
		t.Fatalf("expected formatted price, got %q", data.Price)
	}
}

// Testa parse e validacoes do formulario de plano.
func TestParsePlanForm(t *testing.T) {
	tests := []struct {
		name      string
		values    url.Values
		wantError bool
	}{
		{
			name: "valid",
			values: url.Values{
				"name":          {"Plano"},
				"duration_days": {"30"},
				"price":         {"9,90"},
				"description":   {"desc"},
				"active":        {"on"},
			},
			wantError: false,
		},
		{
			name: "missing-name",
			values: url.Values{
				"duration_days": {"30"},
				"price":         {"9,90"},
			},
			wantError: true,
		},
		{
			name: "invalid-duration",
			values: url.Values{
				"name":          {"Plano"},
				"duration_days": {"0"},
				"price":         {"9,90"},
			},
			wantError: true,
		},
		{
			name: "invalid-price",
			values: url.Values{
				"name":          {"Plano"},
				"duration_days": {"30"},
				"price":         {"x"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, "/plans", strings.NewReader(tt.values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		data := view.PlanFormData{}

		plan, err := parsePlanForm(req, &data)
		if tt.wantError {
			if err == nil {
				t.Fatalf("%s: expected error, got nil", tt.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if plan.Name == "" || plan.DurationDays == 0 {
			t.Fatalf("%s: expected plan fields to be populated", tt.name)
		}
	}
}
