package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/PabloPavan/jaiu/internal/http/handlers"
	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
	"github.com/PabloPavan/jaiu/internal/ports"
)

func New(h *handlers.Handler, sessions ports.SessionStore, cookieName string, notifyCfg httpmw.NotifyConfig, eventHandler http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmw.Observability())
	r.Use(httpmw.Notify(notifyCfg))
	r.Use(middleware.Recoverer)

	r.Get("/healthz", h.Health)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", h.Login)
		r.Post("/login", h.LoginPost)
		r.Post("/logout", h.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(httpmw.RequireSession(sessions, cookieName))

		if eventHandler != nil {
			r.Get("/events", eventHandler.ServeHTTP)
		}

		r.Get("/", h.Home)

		//	r.Mount("/students", student.Routes())

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
			r.Get("/new", h.SubscriptionsNew)
			r.Post("/", h.SubscriptionsCreate)
			r.Get("/{subscriptionID}/edit", h.SubscriptionsEdit)
			r.Post("/{subscriptionID}", h.SubscriptionsUpdate)
			r.Post("/{subscriptionID}/cancel", h.SubscriptionsCancel)
		})

		r.Route("/payments", func(r chi.Router) {
			r.Get("/", h.PaymentsIndex)
			r.Get("/new", h.PaymentsNew)
			r.Post("/", h.PaymentsCreate)
			r.Get("/{paymentID}/edit", h.PaymentsEdit)
			r.Post("/{paymentID}", h.PaymentsUpdate)
			r.Post("/{paymentID}/reverse", h.PaymentsReverse)
		})

		r.Route("/reports", func(r chi.Router) {
			r.Get("/", h.ReportsIndex)
		})
	})

	if imageHandler := h.ImageHandler(); imageHandler != nil {
		r.Handle("/images/*", http.StripPrefix("/images", imageHandler))
	}
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("web/static"))))

	return r
}
