package handlers

import "net/http"

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, page("Entrar", "page/auth/login", nil))
}
