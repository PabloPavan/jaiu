package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/internal/view"
)

type Handler struct {
	renderer *view.Renderer
}

func New(renderer *view.Renderer) *Handler {
	return &Handler{renderer: renderer}
}

func (h *Handler) renderPage(w http.ResponseWriter, data view.PageData) {
	data.Now = time.Now()
	if err := h.renderer.Render(w, data); err != nil {
		log.Printf("render error: %v", err)
	}
}

func (h *Handler) renderPartial(w http.ResponseWriter, name string, data any) {
	if err := h.renderer.RenderPartial(w, name, data); err != nil {
		log.Printf("render partial error: %v", err)
	}
}
