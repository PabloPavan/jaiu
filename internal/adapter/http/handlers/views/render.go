package view

import (
	"net/http"

	"github.com/a-h/templ"
)

func RenderPage(w http.ResponseWriter, r *http.Request, page Page) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return BaseLayout(page.Title, page.CurrentUser, page.Content).Render(r.Context(), w)
}

func RenderComponent(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(r.Context(), w)
}
