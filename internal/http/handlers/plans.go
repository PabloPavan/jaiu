package handlers

import "net/http"

func (h *Handler) PlansIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, page("Planos", "page/plans/index", nil))
}
