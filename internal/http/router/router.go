package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/PabloPavan/jaiu/internal/http/handlers"
)

func New(h *handlers.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", h.Health)
	r.Get("/", h.Home)

	r.Route("/students", func(r chi.Router) {
		r.Get("/", h.StudentsIndex)
		r.Get("/preview", h.StudentsPreview)
	})

	r.Route("/plans", func(r chi.Router) {
		r.Get("/", h.PlansIndex)
	})

	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", h.SubscriptionsIndex)
	})

	r.Route("/payments", func(r chi.Router) {
		r.Get("/", h.PaymentsIndex)
	})

	r.Route("/reports", func(r chi.Router) {
		r.Get("/", h.ReportsIndex)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", h.Login)
	})

	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("web/static"))))

	return r
}
