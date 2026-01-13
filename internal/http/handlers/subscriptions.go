package handlers

import "net/http"

func (h *Handler) SubscriptionsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Assinaturas", "page/subscriptions/index", nil))
}
