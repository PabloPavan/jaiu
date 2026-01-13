package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/a-h/templ"
)

type Handler struct {
	auth     AuthService
	plans    PlanService
	sessions ports.SessionStore
	config   SessionConfig
}

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (domain.User, error)
}

type PlanService interface {
	ListActive(ctx context.Context) ([]domain.Plan, error)
}

type SessionConfig struct {
	CookieName string
	TTL        time.Duration
	Secure     bool
	SameSite   http.SameSite
}

func New(auth AuthService, plans PlanService, sessions ports.SessionStore, config SessionConfig) *Handler {
	if config.CookieName == "" {
		config.CookieName = "jaiu_session"
	}
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}
	return &Handler{auth: auth, plans: plans, sessions: sessions, config: config}
}

func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, page view.Page) {
	page.Now = time.Now()
	if session, ok := httpmw.SessionFromContext(r.Context()); ok {
		displayName := session.Name
		if displayName == "" {
			displayName = "Usuario"
		}
		page.CurrentUser = &view.UserInfo{
			Name:        session.Name,
			DisplayName: displayName,
			Role:        string(session.Role),
		}
	}
	if err := view.RenderPage(w, r, page); err != nil {
		log.Printf("render error: %v", err)
	}
}

func (h *Handler) renderComponent(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if err := view.RenderComponent(w, r, component); err != nil {
		log.Printf("render partial error: %v", err)
	}
}
