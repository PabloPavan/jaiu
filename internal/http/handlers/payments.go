package handlers

import "net/http"

func (h *Handler) PaymentsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, page("Pagamentos", "page/payments/index", nil))
}
