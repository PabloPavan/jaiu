package handlers

import (
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/a-h/templ"
)

func page(title string, content templ.Component) view.Page {
	return view.Page{
		Title:   title,
		Content: content,
	}
}
