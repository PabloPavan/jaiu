package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Entrar", view.LoginPage(view.LoginData{})))
}

func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if h.services.Auth == nil {
		http.Error(w, "autenticacao nao configurada", http.StatusNotImplemented)
		return
	}
	if h.sessions == nil {
		http.Error(w, "sessoes nao configuradas", http.StatusNotImplemented)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "erro ao ler formulario", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	viewData := view.LoginData{Email: email}

	if email == "" || password == "" {
		viewData.Error = "Informe email e senha."
		h.renderPage(w, r, page("Entrar", view.LoginPage(viewData)))
		return
	}

	user, err := h.services.Auth.Authenticate(r.Context(), email, password)
	if err != nil {
		if errors.Is(err, ports.ErrUnauthorized) || errors.Is(err, ports.ErrNotFound) {
			viewData.Error = "Credenciais invalidas."
			h.renderPage(w, r, page("Entrar", view.LoginPage(viewData)))
			return
		}
		http.Error(w, "erro ao autenticar", http.StatusInternalServerError)
		return
	}

	session := ports.Session{
		UserID:    user.ID,
		Name:      user.Name,
		Role:      user.Role,
		ExpiresAt: time.Now().Add(h.config.TTL),
	}
	token, err := h.sessions.Create(r.Context(), session)
	if err != nil {
		http.Error(w, "erro ao criar sessao", http.StatusInternalServerError)
		return
	}

	h.recordAuditWithActor(r, user.ID, string(user.Role), "user.login", "user", user.ID, nil)

	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.config.Secure,
		SameSite: h.config.SameSite,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var actor ports.Session
	if h.sessions != nil {
		if cookie, err := r.Cookie(h.config.CookieName); err == nil {
			if session, err := h.sessions.Get(r.Context(), cookie.Value); err == nil {
				actor = session
			}
		}
	}

	if h.sessions != nil {
		if cookie, err := r.Cookie(h.config.CookieName); err == nil {
			_ = h.sessions.Delete(r.Context(), cookie.Value)
		}
	}

	if actor.UserID != "" {
		h.recordAuditWithActor(r, actor.UserID, string(actor.Role), "user.logout", "user", actor.UserID, nil)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.config.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.config.Secure,
		SameSite: h.config.SameSite,
	})

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}
