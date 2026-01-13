package handlers

import "net/http"

func (h *Handler) ReportsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Relatórios", "page/reports/index", nil))
}
