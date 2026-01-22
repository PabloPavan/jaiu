package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PabloPavan/eventrail/sse"
	"github.com/go-chi/chi/v5"
)

type recordingBroker struct {
	lastChannel  string
	lastPayload  []byte
	publishCount int
}

func (b *recordingBroker) Subscribe(ctx context.Context, patterns ...string) (sse.Subscription, error) {
	return nil, errors.New("not implemented")
}

func (b *recordingBroker) Publish(ctx context.Context, channel string, payload []byte) error {
	b.lastChannel = channel
	b.lastPayload = append([]byte(nil), payload...)
	b.publishCount++
	return nil
}

// Testa que o middleware publica evento com payload completo em sucesso.
func TestNotify_PublishesEvent(t *testing.T) {
	broker := &recordingBroker{}
	publisher := sse.NewPublisher(broker)

	router := chi.NewRouter()
	router.Use(Notify(NotifyConfig{
		Publisher: publisher,
		Scope:     "tenant",
		ScopeID:   42,
		Context:   context.Background(),
	}))
	router.Post("/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/students/123?origin_id=query", nil)
	req.Header.Set(originHeader, "header-origin")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if broker.publishCount != 1 {
		t.Fatalf("expected 1 publish, got %d", broker.publishCount)
	}
	if broker.lastChannel != "tenant:42:students" {
		t.Fatalf("expected channel tenant:42:students, got %q", broker.lastChannel)
	}

	var event sse.Event
	if err := json.Unmarshal(broker.lastPayload, &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	if event.EventType != "students.changed" {
		t.Fatalf("expected event type students.changed, got %q", event.EventType)
	}

	var payload map[string]any
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got, ok := payload["method"].(string); !ok || got != http.MethodPost {
		t.Fatalf("expected method %q, got %#v", http.MethodPost, payload["method"])
	}
	if got, ok := payload["route"].(string); !ok || got != "/students/{id}" {
		t.Fatalf("expected route %q, got %#v", "/students/{id}", payload["route"])
	}
	if got, ok := payload["path"].(string); !ok || got != "/students/123" {
		t.Fatalf("expected path %q, got %#v", "/students/123", payload["path"])
	}
	if got, ok := payload["origin_id"].(string); !ok || got != "header-origin" {
		t.Fatalf("expected origin_id %q, got %#v", "header-origin", payload["origin_id"])
	}
}

// Testa que o middleware usa o path quando o route pattern nao existe.
func TestNotify_UsesPathWhenRoutePatternMissing(t *testing.T) {
	broker := &recordingBroker{}
	publisher := sse.NewPublisher(broker)

	handler := Notify(NotifyConfig{
		Publisher: publisher,
		Scope:     "tenant",
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/plans", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if broker.publishCount != 1 {
		t.Fatalf("expected 1 publish, got %d", broker.publishCount)
	}
	if broker.lastChannel != "tenant:1:plans" {
		t.Fatalf("expected channel tenant:1:plans, got %q", broker.lastChannel)
	}
}

// Testa que o middleware ignora metodos de leitura.
func TestNotify_SkipsOnReadMethod(t *testing.T) {
	broker := &recordingBroker{}
	publisher := sse.NewPublisher(broker)

	router := chi.NewRouter()
	router.Use(Notify(NotifyConfig{
		Publisher: publisher,
		Scope:     "tenant",
	}))
	router.Get("/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/students/123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if broker.publishCount != 0 {
		t.Fatalf("expected 0 publishes, got %d", broker.publishCount)
	}
}

// Testa que o middleware nao publica quando a resposta e erro.
func TestNotify_SkipsOnErrorStatus(t *testing.T) {
	broker := &recordingBroker{}
	publisher := sse.NewPublisher(broker)

	router := chi.NewRouter()
	router.Use(Notify(NotifyConfig{
		Publisher: publisher,
		Scope:     "tenant",
	}))
	router.Post("/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodPost, "/students/123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if broker.publishCount != 0 {
		t.Fatalf("expected 0 publishes, got %d", broker.publishCount)
	}
}

// Testa a matriz de metodos que disparam notificacao.
func TestShouldNotifyMethod(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, true},
		{http.MethodGet, false},
		{http.MethodHead, false},
	}

	for _, tt := range tests {
		if got := shouldNotifyMethod(tt.method); got != tt.want {
			t.Fatalf("shouldNotifyMethod(%q) = %v, want %v", tt.method, got, tt.want)
		}
	}
}

// Testa a extracao do topico a partir da rota.
func TestTopicFromRoute(t *testing.T) {
	tests := []struct {
		route string
		want  string
	}{
		{"", ""},
		{"/", "home"},
		{"/students/{id}", "students"},
		{"  /plans/{id}?foo=bar  ", "plans"},
		{"/static/app.js", ""},
		{"/images/logo.png", ""},
		{"/events", ""},
		{"/healthz", ""},
		{"/Payments/{id}", "payments"},
	}

	for _, tt := range tests {
		if got := topicFromRoute(tt.route); got != tt.want {
			t.Fatalf("topicFromRoute(%q) = %q, want %q", tt.route, got, tt.want)
		}
	}
}
