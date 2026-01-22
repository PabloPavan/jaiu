// package router

// import (
// 	"context"
// 	"mime/multipart"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"strings"
// 	"testing"

// 	"github.com/PabloPavan/jaiu/internal/domain"
// 	"github.com/PabloPavan/jaiu/internal/http/handlers"
// 	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
// 	"github.com/PabloPavan/jaiu/internal/ports"
// )

// type fakeSessionStore struct {
// 	sessions    map[string]ports.Session
// 	lastCreated ports.Session
// 	createToken string
// 	createErr   error
// 	getErr      error
// 	deleteErr   error
// }

// func (s *fakeSessionStore) Create(ctx context.Context, session ports.Session) (string, error) {
// 	if s.createErr != nil {
// 		return "", s.createErr
// 	}
// 	if s.sessions == nil {
// 		s.sessions = make(map[string]ports.Session)
// 	}
// 	s.lastCreated = session
// 	token := s.createToken
// 	if token == "" {
// 		token = "token-1"
// 	}
// 	s.sessions[token] = session
// 	return token, nil
// }

// func (s *fakeSessionStore) Get(ctx context.Context, token string) (ports.Session, error) {
// 	if s.getErr != nil {
// 		return ports.Session{}, s.getErr
// 	}
// 	session, ok := s.sessions[token]
// 	if !ok {
// 		return ports.Session{}, ports.ErrNotFound
// 	}
// 	return session, nil
// }

// func (s *fakeSessionStore) Delete(ctx context.Context, token string) error {
// 	if s.deleteErr != nil {
// 		return s.deleteErr
// 	}
// 	delete(s.sessions, token)
// 	return nil
// }

// type fakeAuthService struct {
// 	user        domain.User
// 	err         error
// 	gotEmail    string
// 	gotPassword string
// }

// func (a *fakeAuthService) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
// 	a.gotEmail = email
// 	a.gotPassword = password
// 	if a.err != nil {
// 		return domain.User{}, a.err
// 	}
// 	return a.user, nil
// }

// func (a *fakeAuthService) Logout(ctx context.Context, session ports.Session) error {
// 	return nil
// }

// type fakeImageService struct {
// 	handler http.Handler
// }

// func (f *fakeImageService) UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
// 	return "", nil
// }

// func (f *fakeImageService) Handler() http.Handler {
// 	return f.handler
// }

// // Testa que o healthcheck responde sem depender de sessao.
// func TestRouterHealthz(t *testing.T) {
// 	h := handlers.New(handlers.Services{}, nil, handlers.SessionConfig{})
// 	r := New(h, nil, "", httpmw.NotifyConfig{}, nil)

// 	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
// 	rec := httptest.NewRecorder()
// 	r.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", rec.Code)
// 	}
// 	if strings.TrimSpace(rec.Body.String()) != "ok" {
// 		t.Fatalf("expected body ok, got %q", rec.Body.String())
// 	}
// }

// // Testa que rotas protegidas redirecionam sem cookie de sessao.
// func TestRouterProtectedRedirectsWithoutSession(t *testing.T) {
// 	store := &fakeSessionStore{}
// 	h := handlers.New(handlers.Services{}, store, handlers.SessionConfig{CookieName: "test_session"})
// 	r := New(h, store, "test_session", httpmw.NotifyConfig{}, nil)

// 	req := httptest.NewRequest(http.MethodGet, "/students/", nil)
// 	rec := httptest.NewRecorder()
// 	r.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusSeeOther {
// 		t.Fatalf("expected 303, got %d", rec.Code)
// 	}
// 	if rec.Header().Get("Location") != "/auth/login" {
// 		t.Fatalf("expected redirect to /auth/login, got %q", rec.Header().Get("Location"))
// 	}
// }

// // Testa que rotas protegidas liberam acesso com cookie valido.
// func TestRouterProtectedAllowsWithSession(t *testing.T) {
// 	store := &fakeSessionStore{
// 		sessions: map[string]ports.Session{
// 			"token-1": {UserID: "user-1", Role: domain.RoleAdmin},
// 		},
// 	}
// 	h := handlers.New(handlers.Services{}, store, handlers.SessionConfig{CookieName: "test_session"})
// 	r := New(h, store, "test_session", httpmw.NotifyConfig{}, nil)

// 	req := httptest.NewRequest(http.MethodGet, "/students/", nil)
// 	req.AddCookie(&http.Cookie{Name: "test_session", Value: "token-1"})
// 	rec := httptest.NewRecorder()
// 	r.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("expected 200, got %d", rec.Code)
// 	}
// }

// // Testa que a rota /events so existe quando o handler e fornecido.
// func TestRouterEventsRouteOptional(t *testing.T) {
// 	store := &fakeSessionStore{
// 		sessions: map[string]ports.Session{
// 			"token-1": {UserID: "user-1", Role: domain.RoleAdmin},
// 		},
// 	}
// 	h := handlers.New(handlers.Services{}, store, handlers.SessionConfig{CookieName: "test_session"})

// 	noEvents := New(h, store, "test_session", httpmw.NotifyConfig{}, nil)
// 	req := httptest.NewRequest(http.MethodGet, "/events", nil)
// 	req.AddCookie(&http.Cookie{Name: "test_session", Value: "token-1"})
// 	rec := httptest.NewRecorder()
// 	noEvents.ServeHTTP(rec, req)
// 	if rec.Code != http.StatusNotFound {
// 		t.Fatalf("expected 404 without events handler, got %d", rec.Code)
// 	}

// 	eventHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusNoContent)
// 	})
// 	withEvents := New(h, store, "test_session", httpmw.NotifyConfig{}, eventHandler)

// 	unauthReq := httptest.NewRequest(http.MethodGet, "/events", nil)
// 	unauthRec := httptest.NewRecorder()
// 	withEvents.ServeHTTP(unauthRec, unauthReq)
// 	if unauthRec.Code != http.StatusSeeOther {
// 		t.Fatalf("expected 303 without session, got %d", unauthRec.Code)
// 	}

// 	authReq := httptest.NewRequest(http.MethodGet, "/events", nil)
// 	authReq.AddCookie(&http.Cookie{Name: "test_session", Value: "token-1"})
// 	authRec := httptest.NewRecorder()
// 	withEvents.ServeHTTP(authRec, authReq)
// 	if authRec.Code != http.StatusNoContent {
// 		t.Fatalf("expected 204 with events handler, got %d", authRec.Code)
// 	}
// }

// // Testa que a rota /images depende da configuracao do ImageService.
// func TestRouterImagesRouteOptional(t *testing.T) {
// 	store := &fakeSessionStore{}
// 	h := handlers.New(handlers.Services{}, store, handlers.SessionConfig{CookieName: "test_session"})

// 	withoutImages := New(h, store, "test_session", httpmw.NotifyConfig{}, nil)
// 	req := httptest.NewRequest(http.MethodGet, "/images/test", nil)
// 	rec := httptest.NewRecorder()
// 	withoutImages.ServeHTTP(rec, req)
// 	if rec.Code != http.StatusNotFound {
// 		t.Fatalf("expected 404 without image handler, got %d", rec.Code)
// 	}

// 	h.SetImageConfig(handlers.ImageConfig{
// 		ImageService: &fakeImageService{handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			w.WriteHeader(http.StatusAccepted)
// 		})},
// 	})
// 	withImages := New(h, store, "test_session", httpmw.NotifyConfig{}, nil)
// 	imgReq := httptest.NewRequest(http.MethodGet, "/images/test", nil)
// 	imgRec := httptest.NewRecorder()
// 	withImages.ServeHTTP(imgRec, imgReq)
// 	if imgRec.Code != http.StatusAccepted {
// 		t.Fatalf("expected 202 with image handler, got %d", imgRec.Code)
// 	}
// }

// // Testa o fluxo de login via POST gerando cookie e redirecionamento.
// func TestRouterLoginPost(t *testing.T) {
// 	auth := &fakeAuthService{
// 		user: domain.User{ID: "user-1", Name: "User", Role: domain.RoleAdmin, Active: true},
// 	}
// 	store := &fakeSessionStore{}
// 	h := handlers.New(handlers.Services{Auth: auth}, store, handlers.SessionConfig{CookieName: "test_session"})
// 	r := New(h, store, "test_session", httpmw.NotifyConfig{}, nil)

// 	values := url.Values{
// 		"email":    {"user@example.com"},
// 		"password": {"secret"},
// 	}
// 	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(values.Encode()))
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	rec := httptest.NewRecorder()
// 	r.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusSeeOther {
// 		t.Fatalf("expected 303, got %d", rec.Code)
// 	}
// 	if rec.Header().Get("Location") != "/" {
// 		t.Fatalf("expected redirect to /, got %q", rec.Header().Get("Location"))
// 	}
// 	if !strings.Contains(rec.Header().Get("Set-Cookie"), "test_session=") {
// 		t.Fatalf("expected session cookie, got %q", rec.Header().Get("Set-Cookie"))
// 	}
// 	if store.lastCreated.UserID != "user-1" {
// 		t.Fatalf("expected session for user-1, got %#v", store.lastCreated)
// 	}
// 	if auth.gotEmail != "user@example.com" || auth.gotPassword != "secret" {
// 		t.Fatalf("expected auth to receive credentials, got %q/%q", auth.gotEmail, auth.gotPassword)
// 	}
// }
