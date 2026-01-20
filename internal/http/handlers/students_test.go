package handlers

import (
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

// Testa o parse do status do aluno com default e invalidacao.
func TestParseStudentStatus(t *testing.T) {
	if got, err := parseStudentStatus(""); err != nil || got != domain.StudentActive {
		t.Fatalf("expected default active, got %q err=%v", got, err)
	}
	if _, err := parseStudentStatus("invalid"); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// Testa o filtro de status do aluno para a busca.
func TestStatusFilter(t *testing.T) {
	if got := statusFilter("all"); got != nil {
		t.Fatal("expected nil for all")
	}
	if got := statusFilter(string(domain.StudentInactive)); len(got) != 1 || got[0] != domain.StudentInactive {
		t.Fatalf("expected inactive slice, got %#v", got)
	}
}

// Testa normalizacao do valor de status do aluno.
func TestNormalizeStudentStatusValue(t *testing.T) {
	if got := normalizeStudentStatusValue("all"); got != "all" {
		t.Fatalf("expected all, got %q", got)
	}
	if got := normalizeStudentStatusValue("unknown"); got != string(domain.StudentActive) {
		t.Fatalf("expected active fallback, got %q", got)
	}
}

// Testa apresentacao do status do aluno.
func TestStatusPresentation(t *testing.T) {
	label := statusPresentation(domain.StudentSuspended)
	if label != "Suspenso" {
		t.Fatalf("expected Suspenso, got %q", label)
	}
	label = statusPresentation(domain.StudentActive)
	if label != "Ativo" {
		t.Fatalf("expected Ativo, got %q", label)
	}
}

// Testa calculo de iniciais do aluno.
func TestStudentInitials(t *testing.T) {
	if got := studentInitials("Maria Silva"); got != "MS" {
		t.Fatalf("expected MS, got %q", got)
	}
	if got := studentInitials("  "); got != "" {
		t.Fatalf("expected empty initials, got %q", got)
	}
}

// Testa inicial a partir de uma palavra.
func TestInitialFrom(t *testing.T) {
	if got := initialFrom("ana"); got != "A" {
		t.Fatalf("expected A, got %q", got)
	}
	if got := initialFrom(" "); got != "" {
		t.Fatalf("expected empty initial, got %q", got)
	}
}

// Testa formatacao de data opcional do aluno.
func TestFormatDateBR(t *testing.T) {
	if got := formatDateBR(nil); got != "" {
		t.Fatalf("expected empty string for nil date, got %q", got)
	}
	date := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	if got := formatDateBR(&date); got != "01/05/2024" {
		t.Fatalf("expected 01/05/2024, got %q", got)
	}
}

// Testa parse de datas em formatos diferentes.
func TestParseDateInput(t *testing.T) {
	if got, err := parseDateInput(""); err != nil || got != nil {
		t.Fatalf("expected nil nil, got %v err=%v", got, err)
	}
	if _, err := parseDateInput("99/99/9999"); err == nil {
		t.Fatal("expected error for invalid date")
	}
	if got, err := parseDateInput("2024-05-01"); err != nil || got == nil || got.Format("2006-01-02") != "2024-05-01" {
		t.Fatalf("expected parsed date, got %v err=%v", got, err)
	}
	if got, err := parseDateInput("01/05/2024"); err != nil || got == nil || got.Format("02/01/2006") != "01/05/2024" {
		t.Fatalf("expected parsed date, got %v err=%v", got, err)
	}
}

// Testa formatacao de data para input BR.
func TestFormatDateInputBR(t *testing.T) {
	if got := formatDateInputBR(nil); got != "" {
		t.Fatalf("expected empty string for nil date, got %q", got)
	}
	date := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	if got := formatDateInputBR(&date); got != "01/05/2024" {
		t.Fatalf("expected 01/05/2024, got %q", got)
	}
}

// Testa montagem da URL da foto do aluno.
func TestPhotoURLForVariant(t *testing.T) {
	h := &Handler{}
	if got := h.photoURLForVariant("", "preview"); got != "" {
		t.Fatalf("expected empty for missing key, got %q", got)
	}
	h.images.BaseURL = "http://cdn/"
	if got := h.photoURLForVariant("object", "list"); got != "http://cdn/object/list" {
		t.Fatalf("expected normalized url, got %q", got)
	}
}
