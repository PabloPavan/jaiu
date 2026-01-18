package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testa que o statusRecorder grava status e bytes quando WriteHeader e chamado.
func TestStatusRecorder_WriteHeader(t *testing.T) {
	base := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: base}

	rec.WriteHeader(http.StatusAccepted)
	rec.Write([]byte("ok"))

	if rec.status != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.status)
	}
	if rec.bytes != 2 {
		t.Fatalf("expected 2 bytes, got %d", rec.bytes)
	}
}

// Testa que o statusRecorder assume 200 quando Write eh chamado direto.
func TestStatusRecorder_WriteDefaults(t *testing.T) {
	base := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: base}

	rec.Write([]byte("ok"))

	if rec.status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.status)
	}
	if rec.bytes != 2 {
		t.Fatalf("expected 2 bytes, got %d", rec.bytes)
	}
}

// Testa que o notifyRecorder preserva o primeiro status escrito.
func TestNotifyRecorder_PreservesFirstStatus(t *testing.T) {
	base := httptest.NewRecorder()
	rec := &notifyRecorder{ResponseWriter: base, status: http.StatusOK}

	rec.WriteHeader(http.StatusCreated)
	rec.WriteHeader(http.StatusBadRequest)

	if rec.status != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.status)
	}
	if !rec.wroteHeader {
		t.Fatal("expected wroteHeader to be true")
	}
}

// Testa que o notifyRecorder seta 200 e marca header quando Write e chamado.
func TestNotifyRecorder_WriteDefaults(t *testing.T) {
	base := httptest.NewRecorder()
	rec := &notifyRecorder{ResponseWriter: base, status: http.StatusOK}

	rec.Write([]byte("ok"))

	if rec.status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.status)
	}
	if !rec.wroteHeader {
		t.Fatal("expected wroteHeader to be true")
	}
}
