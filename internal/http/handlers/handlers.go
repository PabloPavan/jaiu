package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	httpmw "github.com/PabloPavan/jaiu/internal/http/middleware"
	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/a-h/templ"
)

type Handler struct {
	services Services
	sessions ports.SessionStore
	config   SessionConfig
	images   ImageConfig
}

type Services struct {
	Auth          AuthService
	Plans         PlanService
	Students      StudentService
	Subscriptions SubscriptionService
	Payments      PaymentService
}

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (domain.User, error)
	Logout(ctx context.Context, session ports.Session) error
}

type PlanService interface {
	ListActive(ctx context.Context) ([]domain.Plan, error)
	FindByID(ctx context.Context, id string) (domain.Plan, error)
	Create(ctx context.Context, plan domain.Plan) (domain.Plan, error)
	Update(ctx context.Context, plan domain.Plan) (domain.Plan, error)
	Deactivate(ctx context.Context, planID string) (domain.Plan, error)
}

type StudentService interface {
	Count(ctx context.Context, filter ports.StudentFilter) (int, error)
	Search(ctx context.Context, filter ports.StudentFilter) ([]domain.Student, error)
	FindByID(ctx context.Context, id string) (domain.Student, error)
	Register(ctx context.Context, student domain.Student) (domain.Student, error)
	Update(ctx context.Context, student domain.Student) (domain.Student, error)
	SetStatus(ctx context.Context, studentID string, status domain.StudentStatus) (domain.Student, error)
	Deactivate(ctx context.Context, studentID string) (domain.Student, error)
}

type SubscriptionService interface {
	FindByID(ctx context.Context, id string) (domain.Subscription, error)
	Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	Cancel(ctx context.Context, subscriptionID string) (domain.Subscription, error)
	ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error)
	DueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error)
}

type PaymentService interface {
	FindByID(ctx context.Context, id string) (domain.Payment, error)
	Register(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	Update(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	Reverse(ctx context.Context, paymentID string) (domain.Payment, error)
	ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error)
	ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error)
}

type SessionConfig struct {
	CookieName string
	TTL        time.Duration
	Secure     bool
	SameSite   http.SameSite
}

func New(services Services, sessions ports.SessionStore, config SessionConfig) *Handler {
	if config.CookieName == "" {
		config.CookieName = "jaiu_session"
	}
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}
	return &Handler{services: services, sessions: sessions, config: config}
}

func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, page view.Page) {
	page.Now = time.Now()
	if session, ok := httpmw.SessionFromContext(r.Context()); ok {
		displayName := session.Name
		if displayName == "" {
			displayName = "Usuario"
		}
		page.CurrentUser = &view.UserInfo{
			Name:        session.Name,
			DisplayName: displayName,
			Role:        string(session.Role),
		}
	}
	if err := view.RenderPage(w, r, page); err != nil {
		observability.Logger(r.Context()).Error("failed to render page", "err", err)
	}
}

func (h *Handler) renderComponent(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if err := view.RenderComponent(w, r, component); err != nil {
		observability.Logger(r.Context()).Error("failed to render component", "err", err)
	}
}

func (h *Handler) renderFormError(w http.ResponseWriter, r *http.Request, title string, component templ.Component) {
	if isHTMX(r) {
		h.renderComponent(w, r, component)
		return
	}
	h.renderPage(w, r, page(title, component))
}

func (h *Handler) renderHTMXOrPage(w http.ResponseWriter, r *http.Request, title string, pageComponent, htmxComponent templ.Component) {
	if isHTMX(r) {
		h.renderComponent(w, r, htmxComponent)
		return
	}
	h.renderPage(w, r, page(title, pageComponent))
}

func (h *Handler) renderHTMXOrRedirect(w http.ResponseWriter, r *http.Request, location string, render func()) {
	if isHTMX(r) {
		render()
		return
	}
	http.Redirect(w, r, location, http.StatusSeeOther)
}

func (h *Handler) redirectHTMXOrRedirect(w http.ResponseWriter, r *http.Request, location string) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", location)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, location, http.StatusSeeOther)
}
