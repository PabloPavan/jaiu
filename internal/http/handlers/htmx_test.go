package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testa o header HX-Request para identificar requisicoes HTMX.
func TestIsHTMX(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if isHTMX(req) {
		t.Fatal("expected non-htmx when header is missing")
	}

	req.Header.Set("HX-Request", "true")
	if !isHTMX(req) {
		t.Fatal("expected htmx when header is true")
	}
}
