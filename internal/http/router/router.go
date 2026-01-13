package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/PabloPavan/jaiu/internal/http/handlers"
	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
	"github.com/PabloPavan/jaiu/internal/ports"
)

func New(h *handlers.Handler, sessions ports.SessionStore, cookieName string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", h.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", h.Login)
		r.Post("/login", h.LoginPost)
		r.Post("/logout", h.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(httpmw.RequireSession(sessions, cookieName))

		r.Get("/", h.Home)

		r.Route("/students", func(r chi.Router) {
			r.Get("/", h.StudentsIndex)
			r.Get("/new", h.StudentsNew)
			r.Post("/", h.StudentsCreate)
			r.Get("/{studentID}/edit", h.StudentsEdit)
			r.Post("/{studentID}", h.StudentsUpdate)
			r.Post("/{studentID}/delete", h.StudentsDelete)
			r.With(httpmw.RequireHTMX).Get("/preview", h.StudentsPreview)
		})

		r.Route("/plans", func(r chi.Router) {
			r.Get("/", h.PlansIndex)
			r.Get("/new", h.PlansNew)
			r.Post("/", h.PlansCreate)
			r.Get("/{planID}/edit", h.PlansEdit)
			r.Post("/{planID}", h.PlansUpdate)
			r.Post("/{planID}/delete", h.PlansDelete)
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
	})

	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("web/static"))))

	return r
}
