package handlers

import (
	"net/http"

	"github.com/PabloPavan/jaiu/internal/view"
)

func (h *Handler) StudentsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Alunos", view.StudentsPage()))
}

func (h *Handler) StudentsPreview(w http.ResponseWriter, r *http.Request) {
	data := view.StudentsPreviewData{
		Items: []string{
			"Ana Souza · Ativa · Renovar em 10 dias",
			"Carlos Lima · Inativo · Último pagamento há 45 dias",
			"Juliana Costa · Ativa · Mensal",
		},
	}

	h.renderComponent(w, r, view.StudentsPreview(data))
}
