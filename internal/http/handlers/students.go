package handlers

import "net/http"

type studentPreviewData struct {
	Items []string
}

func (h *Handler) StudentsIndex(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, page("Alunos", "page/students/index", nil))
}

func (h *Handler) StudentsPreview(w http.ResponseWriter, r *http.Request) {
	data := studentPreviewData{
		Items: []string{
			"Ana Souza · Ativa · Renovar em 10 dias",
			"Carlos Lima · Inativo · Último pagamento há 45 dias",
			"Juliana Costa · Ativa · Mensal",
		},
	}

	h.renderPartial(w, "partial/students-preview", data)
}
