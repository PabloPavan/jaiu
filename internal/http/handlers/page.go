package handlers

import "github.com/PabloPavan/jaiu/internal/view"

func page(title, template string, data any) view.PageData {
	return view.PageData{
		Title:           title,
		ContentTemplate: template,
		Data:            data,
	}
}
