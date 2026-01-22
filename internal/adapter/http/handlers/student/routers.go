package handlers

import (
	httpmw "github.com/PabloPavan/jaiu/internal/adapter/http/middleware"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.StudentsIndex)
	r.Get("/new", h.StudentsNew)
	r.Post("/", h.StudentsCreate)
	r.Get("/{studentID}/edit", h.StudentsEdit)
	r.Post("/{studentID}", h.StudentsUpdate)
	r.Post("/{studentID}/delete", h.StudentsDelete)
	r.With(httpmw.RequireHTMX).Get("/preview", h.StudentsPreview)

	return r
}
