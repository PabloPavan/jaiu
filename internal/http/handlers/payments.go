package handlers

import (
	"net/http"

	"github.com/PabloPavan/jaiu/internal/view"
)

func (h *Handler) PaymentsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Pagamentos", view.PaymentsPage()))
}
