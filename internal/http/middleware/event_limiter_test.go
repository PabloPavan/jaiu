package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEventLimiterCancelsPreviousConnection(t *testing.T) {
	limiter := NewEventLimiter("test_session")
	started := make(chan string, 2)
	canceled := make(chan string, 2)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Req-ID")
		started <- id
		<-r.Context().Done()
		canceled <- id
	})

	srv := limiter.Wrap(handler)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	req1 := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx1)
	req1.AddCookie(&http.Cookie{Name: "test_session", Value: "token-1"})
	req1.Header.Set("X-Req-ID", "1")
	rec1 := httptest.NewRecorder()
	go srv.ServeHTTP(rec1, req1)

	select {
	case id := <-started:
		if id != "1" {
			t.Fatalf("expected first request start, got %q", id)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for first request to start")
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	req2 := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx2)
	req2.AddCookie(&http.Cookie{Name: "test_session", Value: "token-1"})
	req2.Header.Set("X-Req-ID", "2")
	rec2 := httptest.NewRecorder()
	go srv.ServeHTTP(rec2, req2)

	select {
	case id := <-started:
		if id != "2" {
			t.Fatalf("expected second request start, got %q", id)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for second request to start")
	}

	select {
	case id := <-canceled:
		if id != "1" {
			t.Fatalf("expected first request to be canceled, got %q", id)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for first request to be canceled")
	}

	cancel2()
	select {
	case id := <-canceled:
		if id != "2" {
			t.Fatalf("expected second request to be canceled, got %q", id)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for second request to be canceled")
	}
}

func TestEventLimiterAllowsDifferentSessions(t *testing.T) {
	limiter := NewEventLimiter("test_session")
	started := make(chan string, 2)
	canceled := make(chan string, 2)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Req-ID")
		started <- id
		<-r.Context().Done()
		canceled <- id
	})

	srv := limiter.Wrap(handler)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	req1 := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx1)
	req1.AddCookie(&http.Cookie{Name: "test_session", Value: "token-1"})
	req1.Header.Set("X-Req-ID", "1")
	rec1 := httptest.NewRecorder()
	go srv.ServeHTTP(rec1, req1)

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	req2 := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx2)
	req2.AddCookie(&http.Cookie{Name: "test_session", Value: "token-2"})
	req2.Header.Set("X-Req-ID", "2")
	rec2 := httptest.NewRecorder()
	go srv.ServeHTTP(rec2, req2)

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for requests to start")
		}
	}

	select {
	case id := <-canceled:
		t.Fatalf("did not expect cancellation, got %q", id)
	case <-time.After(150 * time.Millisecond):
	}

	cancel1()
	cancel2()

	for i := 0; i < 2; i++ {
		select {
		case <-canceled:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for requests to cancel")
		}
	}
}
