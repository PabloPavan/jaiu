package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testa que RequireHTMX bloqueia requisicoes sem header HX-Request.
func TestRequireHTMXRejectsNonHTMX(t *testing.T) {
	handler := RequireHTMX(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// Testa que RequireHTMX permite requisicoes HTMX.
func TestRequireHTMXAllows(t *testing.T) {
	called := false
	handler := RequireHTMX(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
