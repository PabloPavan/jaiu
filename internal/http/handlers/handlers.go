package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
)

type Handler struct {
	renderer *view.Renderer
	auth     AuthService
	sessions ports.SessionStore
	config   SessionConfig
}

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (domain.User, error)
}

type SessionConfig struct {
	CookieName string
	TTL        time.Duration
	Secure     bool
	SameSite   http.SameSite
}

func New(renderer *view.Renderer, auth AuthService, sessions ports.SessionStore, config SessionConfig) *Handler {
	if config.CookieName == "" {
		config.CookieName = "jaiu_session"
	}
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}
	return &Handler{renderer: renderer, auth: auth, sessions: sessions, config: config}
}

func (h *Handler) renderPage(w http.ResponseWriter, data view.PageData) {
	data.Now = time.Now()
	if err := h.renderer.Render(w, data); err != nil {
		log.Printf("render error: %v", err)
	}
}

func (h *Handler) renderPartial(w http.ResponseWriter, name string, data any) {
	if err := h.renderer.RenderPartial(w, name, data); err != nil {
		log.Printf("render partial error: %v", err)
	}
}
