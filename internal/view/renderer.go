package view

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Renderer struct {
	templates *template.Template
}

type PageData struct {
	Title           string
	ContentTemplate string
	Data            any
	Now             time.Time
}

func NewRenderer() (*Renderer, error) {
	root, err := templateRoot()
	if err != nil {
		return nil, err
	}

	templates, err := parseTemplates(root)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return &Renderer{templates: templates}, nil
}

func parseTemplates(root string) (*template.Template, error) {
	fs := os.DirFS(root)

	patterns := []string{
		"layouts/*.tmpl",
		"partials/*.tmpl",
		"pages/*.tmpl",
		"pages/*/*.tmpl",
	}

	return template.New("").Funcs(template.FuncMap{
		"asset": func(path string) string {
			return filepath.ToSlash(path)
		},
	}).ParseFS(fs, patterns...)
}

func templateRoot() (string, error) {
	candidates := []string{
		"web/templates",
		"templates",
	}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "web", "templates"))
	}

	for _, path := range candidates {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path, nil
		}
	}

	return "", fmt.Errorf("templates dir not found; tried %v", candidates)
}

func (r *Renderer) Render(w http.ResponseWriter, data PageData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if data.ContentTemplate == "" {
		data.ContentTemplate = "page/home"
	}

	pageTemplate, err := r.templates.Clone()
	if err != nil {
		http.Error(w, "erro ao renderizar template", http.StatusInternalServerError)
		return err
	}

	content := fmt.Sprintf(`{{ define "content" }}{{ template %q . }}{{ end }}`, data.ContentTemplate)
	if _, err := pageTemplate.Parse(content); err != nil {
		http.Error(w, "erro ao renderizar template", http.StatusInternalServerError)
		return err
	}

	if err := pageTemplate.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "erro ao renderizar template", http.StatusInternalServerError)
		return err
	}

	return nil
}

func (r *Renderer) RenderPartial(w http.ResponseWriter, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	partialTemplate, err := r.templates.Clone()
	if err != nil {
		http.Error(w, "erro ao renderizar template", http.StatusInternalServerError)
		return err
	}

	if err := partialTemplate.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "erro ao renderizar template", http.StatusInternalServerError)
		return err
	}

	return nil
}
