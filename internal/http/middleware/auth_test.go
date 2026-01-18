package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PabloPavan/jaiu/internal/auditctx"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type fakeSessionStore struct {
	session  ports.Session
	err      error
	gotToken string
}

func (s *fakeSessionStore) Create(ctx context.Context, session ports.Session) (string, error) {
	return "", nil
}

func (s *fakeSessionStore) Get(ctx context.Context, token string) (ports.Session, error) {
	s.gotToken = token
	if s.err != nil {
		return ports.Session{}, s.err
	}
	return s.session, nil
}

func (s *fakeSessionStore) Delete(ctx context.Context, token string) error {
	return nil
}

// Testa que retorna 501 quando nao ha store configurada.
func TestRequireSession_MissingStore(t *testing.T) {
	mw := RequireSession(nil, "session")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	rec := httptest.NewRecorder()

	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, rec.Code)
	}
}

// Testa negacao quando o cookie de sessao esta ausente.
func TestRequireSession_DenyOnMissingCookie(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "get-redirect",
			method:       http.MethodGet,
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/auth/login",
		},
		{
			name:       "post-unauthorized",
			method:     http.MethodPost,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		store := &fakeSessionStore{}
		mw := RequireSession(store, "session")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})

		req := httptest.NewRequest(tt.method, "/private", nil)
		rec := httptest.NewRecorder()

		mw(handler).ServeHTTP(rec, req)

		if rec.Code != tt.wantStatus {
			t.Fatalf("%s: expected status %d, got %d", tt.name, tt.wantStatus, rec.Code)
		}
		if tt.wantLocation != "" && rec.Header().Get("Location") != tt.wantLocation {
			t.Fatalf("%s: expected location %q, got %q", tt.name, tt.wantLocation, rec.Header().Get("Location"))
		}
	}
}

// Testa negacao quando a store retorna erro (not found vs internal).
func TestRequireSession_DenyOnStoreError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		method     string
		wantStatus int
	}{
		{
			name:       "not-found-post",
			err:        ports.ErrNotFound,
			method:     http.MethodPost,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "internal-error",
			err:        context.Canceled,
			method:     http.MethodPost,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		store := &fakeSessionStore{err: tt.err}
		mw := RequireSession(store, "session")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})

		req := httptest.NewRequest(tt.method, "/private", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "token"})
		rec := httptest.NewRecorder()

		mw(handler).ServeHTTP(rec, req)

		if rec.Code != tt.wantStatus {
			t.Fatalf("%s: expected status %d, got %d", tt.name, tt.wantStatus, rec.Code)
		}
	}
}

// Testa que a sessao valida popula contextos e libera o handler.
func TestRequireSession_AllowsValidSession(t *testing.T) {
	session := ports.Session{
		UserID: "user-1",
		Role:   domain.RoleAdmin,
	}
	store := &fakeSessionStore{session: session}

	mw := RequireSession(store, "session")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := SessionFromContext(r.Context())
		if !ok {
			t.Fatal("expected session in context")
		}
		if got.UserID != session.UserID {
			t.Fatalf("expected userID %q, got %q", session.UserID, got.UserID)
		}

		actor := auditctx.FromContext(r.Context()).Actor
		if actor.ID != session.UserID || actor.Role != string(session.Role) {
			t.Fatalf("unexpected actor in context: %#v", actor)
		}

		activity, ok := userActivityFromContext(r.Context())
		if !ok || activity.UserID != session.UserID || activity.Role != string(session.Role) {
			t.Fatalf("unexpected activity: %#v", activity)
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx, _ := withUserActivity(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/private", nil).WithContext(ctx)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token"})
	rec := httptest.NewRecorder()

	mw(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if store.gotToken != "token" {
		t.Fatalf("expected token %q, got %q", "token", store.gotToken)
	}
}
