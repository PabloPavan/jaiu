package handlers

import "net/http"

func (h *Handler) ReportsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, page("Relatórios", "page/reports/index", nil))
}
