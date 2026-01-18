package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Testa a prioridade de headers/RemoteAddr para descobrir o IP.
func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		{
			name:       "forwarded-for",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1"},
			remoteAddr: "192.0.2.10:1234",
			want:       "203.0.113.1",
		},
		{
			name:       "real-ip",
			headers:    map[string]string{"X-Real-IP": "198.51.100.2"},
			remoteAddr: "192.0.2.10:1234",
			want:       "198.51.100.2",
		},
		{
			name:       "remote-addr-host",
			headers:    nil,
			remoteAddr: "192.0.2.10:1234",
			want:       "192.0.2.10",
		},
		{
			name:       "remote-addr-fallback",
			headers:    nil,
			remoteAddr: "not-a-host",
			want:       "not-a-host",
		},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		req.RemoteAddr = tt.remoteAddr
		for key, value := range tt.headers {
			req.Header.Set(key, value)
		}

		if got := clientIP(req); got != tt.want {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.want, got)
		}
	}
}
