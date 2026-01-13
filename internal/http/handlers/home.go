package handlers

import "net/http"

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Dashboard", "page/home", nil))
}
