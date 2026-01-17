package handlers

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) SubscriptionsIndex(w http.ResponseWriter, r *http.Request) {
	data := h.buildSubscriptionsData(r)
	if isHTMX(r) {
		h.renderComponent(w, r, view.SubscriptionsList(data))
		return
	}
	h.renderPage(w, r, page("Assinaturas", view.SubscriptionsPage(data)))
}

func (h *Handler) SubscriptionsNew(w http.ResponseWriter, r *http.Request) {
	data := h.subscriptionFormCreateData(r)
	h.renderPage(w, r, page("Nova assinatura", view.SubscriptionFormPage(data)))
}

func (h *Handler) SubscriptionsCreate(w http.ResponseWriter, r *http.Request) {
	data := h.subscriptionFormCreateData(r)
	subscription, err := h.parseSubscriptionForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		if isHTMX(r) {
			h.renderComponent(w, r, view.SubscriptionFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
		return
	}

	if h.services.Subscriptions == nil {
		data.Error = "Servico de assinaturas indisponivel."
		if isHTMX(r) {
			h.renderComponent(w, r, view.SubscriptionFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
		return
	}

	_, err = h.services.Subscriptions.Create(r.Context(), subscription)
	if err != nil {
		data.Error = "Nao foi possivel salvar a assinatura."
		if isHTMX(r) {
			h.renderComponent(w, r, view.SubscriptionFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/subscriptions")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

func (h *Handler) SubscriptionsEdit(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionID")
	if h.services.Subscriptions == nil {
		http.NotFound(w, r)
		return
	}

	subscription, err := h.services.Subscriptions.FindByID(r.Context(), subscriptionID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		observability.Logger(r.Context()).Error("failed to load subscription", "err", err)
		http.Error(w, "Erro ao carregar assinatura.", http.StatusInternalServerError)
		return
	}

	data := h.subscriptionFormEditData(r, subscription)
	h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
}

func (h *Handler) SubscriptionsUpdate(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionID")
	data := h.subscriptionFormEditData(r, domain.Subscription{ID: subscriptionID})
	subscription, err := h.parseSubscriptionForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		if isHTMX(r) {
			h.renderComponent(w, r, view.SubscriptionFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
		return
	}

	if h.services.Subscriptions == nil {
		data.Error = "Servico de assinaturas indisponivel."
		if isHTMX(r) {
			h.renderComponent(w, r, view.SubscriptionFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
		return
	}

	subscription.ID = subscriptionID
	_, err = h.services.Subscriptions.Update(r.Context(), subscription)
	if err != nil {
		data.Error = "Nao foi possivel atualizar a assinatura."
		if isHTMX(r) {
			h.renderComponent(w, r, view.SubscriptionFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.SubscriptionFormPage(data)))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/subscriptions")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

func (h *Handler) SubscriptionsCancel(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionID")
	if h.services.Subscriptions == nil {
		http.NotFound(w, r)
		return
	}

	_, err := h.services.Subscriptions.Cancel(r.Context(), subscriptionID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		observability.Logger(r.Context()).Error("failed to cancel subscription", "err", err)
		http.Error(w, "Erro ao cancelar assinatura.", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		data := h.buildSubscriptionsData(r)
		h.renderComponent(w, r, view.SubscriptionsList(data))
		return
	}

	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

func (h *Handler) buildSubscriptionsData(r *http.Request) view.SubscriptionsPageData {
	studentID := strings.TrimSpace(r.FormValue("student_id"))
	status := normalizeSubscriptionStatus(strings.TrimSpace(r.FormValue("status")))

	students := h.loadStudents(r, nil, 500)
	data := view.SubscriptionsPageData{
		StudentID: studentID,
		Status:    status,
		Students:  toStudentOptions(students),
	}

	if h.services.Subscriptions == nil {
		return data
	}

	var subscriptions []domain.Subscription
	if studentID != "" {
		list, err := h.services.Subscriptions.ListByStudent(r.Context(), studentID)
		if err != nil {
			observability.Logger(r.Context()).Error("failed to list subscriptions", "err", err)
			return data
		}
		subscriptions = list
	} else {
		subscriptions = make([]domain.Subscription, 0, len(students))
		for _, student := range students {
			list, err := h.services.Subscriptions.ListByStudent(r.Context(), student.ID)
			if err != nil {
				observability.Logger(r.Context()).Error("failed to list subscriptions", "err", err)
				continue
			}
			subscriptions = append(subscriptions, list...)
		}
	}

	if len(subscriptions) == 0 {
		return data
	}

	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].StartDate.After(subscriptions[j].StartDate)
	})

	planOptions := h.listPlanOptions(r)
	planMap := make(map[string]string, len(planOptions))
	for _, option := range planOptions {
		planMap[option.ID] = option.Name
	}

	studentMap := make(map[string]string, len(students))
	for _, student := range students {
		studentMap[student.ID] = student.FullName
	}

	data.Items = make([]view.SubscriptionItem, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if status != "all" && string(subscription.Status) != status {
			continue
		}

		planName := planMap[subscription.PlanID]
		if planName == "" && h.services.Plans != nil {
			if plan, err := h.services.Plans.FindByID(r.Context(), subscription.PlanID); err == nil {
				planName = plan.Name
			}
		}

		studentName := studentMap[subscription.StudentID]
		if studentName == "" && h.services.Students != nil {
			if student, err := h.services.Students.FindByID(r.Context(), subscription.StudentID); err == nil {
				studentName = student.FullName
			}
		}

		label, className := subscriptionStatusPresentation(subscription.Status)
		item := view.SubscriptionItem{
			ID:          subscription.ID,
			StudentName: studentName,
			PlanName:    planName,
			StartDate:   formatDateBRValue(subscription.StartDate),
			EndDate:     formatDateBRValue(subscription.EndDate),
			PaymentDay:  formatInt(subscription.PaymentDay),
			AutoRenew:   subscription.AutoRenew,
			Status:      string(subscription.Status),
			StatusLabel: label,
			StatusClass: className,
			Price:       formatBRL(subscription.PriceCents),
		}
		data.Items = append(data.Items, item)
	}

	return data
}

func (h *Handler) subscriptionFormCreateData(r *http.Request) view.SubscriptionFormData {
	now := time.Now()
	data := view.SubscriptionFormData{
		Title:       "Nova assinatura",
		Action:      "/subscriptions",
		SubmitLabel: "Criar assinatura",
		Status:      string(domain.SubscriptionActive),
		StartDate:   formatDateBRValue(now),
		PaymentDay:  formatInt(now.Day()),
	}
	h.fillSubscriptionFormOptions(r, &data)
	return data
}

func (h *Handler) subscriptionFormEditData(r *http.Request, subscription domain.Subscription) view.SubscriptionFormData {
	data := view.SubscriptionFormData{
		Title:        "Editar assinatura",
		Action:       "/subscriptions/" + subscription.ID,
		SubmitLabel:  "Salvar",
		DeleteAction: "/subscriptions/" + subscription.ID + "/cancel",
		ShowDelete:   subscription.ID != "",
		StudentID:    subscription.StudentID,
		PlanID:       subscription.PlanID,
		StartDate:    formatDateBRValue(subscription.StartDate),
		EndDate:      formatDateBRValue(subscription.EndDate),
		PaymentDay:   formatInt(subscription.PaymentDay),
		AutoRenew:    subscription.AutoRenew,
		Status:       string(subscription.Status),
		Price:        formatCentsInput(subscription.PriceCents),
	}
	h.fillSubscriptionFormOptions(r, &data)
	data.DisableSelects = true
	return data
}

func (h *Handler) fillSubscriptionFormOptions(r *http.Request, data *view.SubscriptionFormData) {
	data.Students = h.listActiveStudentOptions(r)
	data.Plans = h.listPlanOptions(r)
	if h.services.Students == nil || h.services.Plans == nil {
		data.Error = "Dependencias de assinaturas indisponiveis."
	}

	data.Students = ensureStudentOption(r.Context(), data.Students, data.StudentID, h)
	data.Plans = ensurePlanOption(r.Context(), data.Plans, data.PlanID, h)
}

func (h *Handler) parseSubscriptionForm(r *http.Request, data *view.SubscriptionFormData) (domain.Subscription, error) {
	if err := r.ParseForm(); err != nil {
		return domain.Subscription{}, errors.New("Nao foi possivel ler o formulario.")
	}

	studentID := strings.TrimSpace(r.FormValue("student_id"))
	data.StudentID = studentID
	if studentID == "" {
		return domain.Subscription{}, errors.New("Aluno e obrigatorio.")
	}

	planID := strings.TrimSpace(r.FormValue("plan_id"))
	data.PlanID = planID
	if planID == "" {
		return domain.Subscription{}, errors.New("Plano e obrigatorio.")
	}

	startRaw := strings.TrimSpace(r.FormValue("start_date"))
	data.StartDate = startRaw
	startDate, err := parseDateInput(startRaw)
	if err != nil || startDate == nil {
		return domain.Subscription{}, errors.New("Data de inicio invalida. Use o formato dd/mm/aaaa.")
	}

	endRaw := strings.TrimSpace(r.FormValue("end_date"))
	data.EndDate = endRaw
	endDate, err := parseDateInput(endRaw)
	if err != nil {
		return domain.Subscription{}, errors.New("Vencimento invalido. Use o formato dd/mm/aaaa.")
	}

	paymentDayRaw := strings.TrimSpace(r.FormValue("payment_day"))
	data.PaymentDay = paymentDayRaw
	paymentDay, err := parsePaymentDay(paymentDayRaw)
	if err != nil {
		return domain.Subscription{}, err
	}

	statusValue := strings.TrimSpace(r.FormValue("status"))
	status, err := parseSubscriptionStatus(statusValue)
	if err != nil {
		return domain.Subscription{}, err
	}
	data.Status = string(status)

	priceRaw := strings.TrimSpace(r.FormValue("price"))
	data.Price = priceRaw
	priceCents, err := parsePriceCentsOptional(priceRaw)
	if err != nil {
		return domain.Subscription{}, errors.New("Preco invalido.")
	}

	autoRenew := strings.TrimSpace(r.FormValue("auto_renew")) != ""
	data.AutoRenew = autoRenew

	subscription := domain.Subscription{
		StudentID:  studentID,
		PlanID:     planID,
		StartDate:  *startDate,
		Status:     status,
		PriceCents: priceCents,
		PaymentDay: paymentDay,
		AutoRenew:  autoRenew,
	}
	if endDate != nil {
		subscription.EndDate = *endDate
	}

	return subscription, nil
}

func (h *Handler) listActiveStudentOptions(r *http.Request) []view.StudentOption {
	students := h.loadStudents(r, []domain.StudentStatus{domain.StudentActive}, 500)
	return toStudentOptions(students)
}

func (h *Handler) listStudentOptions(r *http.Request) []view.StudentOption {
	students := h.loadStudents(r, nil, 500)
	return toStudentOptions(students)
}

func (h *Handler) loadStudents(r *http.Request, statuses []domain.StudentStatus, limit int) []domain.Student {
	if h.services.Students == nil {
		return nil
	}

	students, err := h.services.Students.Search(r.Context(), ports.StudentFilter{
		Query:    "",
		Statuses: statuses,
		Limit:    limit,
	})
	if err != nil {
		observability.Logger(r.Context()).Error("failed to list students", "err", err)
		return nil
	}
	return students
}

func toStudentOptions(students []domain.Student) []view.StudentOption {
	options := make([]view.StudentOption, 0, len(students))
	for _, student := range students {
		options = append(options, view.StudentOption{
			ID:   student.ID,
			Name: student.FullName,
		})
	}
	return options
}

func (h *Handler) listPlanOptions(r *http.Request) []view.PlanOption {
	if h.services.Plans == nil {
		return nil
	}

	plans, err := h.services.Plans.ListActive(r.Context())
	if err != nil {
		observability.Logger(r.Context()).Error("failed to list plans", "err", err)
		return nil
	}

	options := make([]view.PlanOption, 0, len(plans))
	for _, plan := range plans {
		options = append(options, view.PlanOption{
			ID:   plan.ID,
			Name: plan.Name,
		})
	}
	return options
}

func ensureStudentOption(ctx context.Context, options []view.StudentOption, studentID string, h *Handler) []view.StudentOption {
	if studentID == "" {
		return options
	}
	for _, option := range options {
		if option.ID == studentID {
			return options
		}
	}
	if h.services.Students == nil {
		return options
	}
	student, err := h.services.Students.FindByID(ctx, studentID)
	if err != nil {
		return options
	}
	return append(options, view.StudentOption{ID: student.ID, Name: student.FullName})
}

func ensurePlanOption(ctx context.Context, options []view.PlanOption, planID string, h *Handler) []view.PlanOption {
	if planID == "" {
		return options
	}
	for _, option := range options {
		if option.ID == planID {
			return options
		}
	}
	if h.services.Plans == nil {
		return options
	}
	plan, err := h.services.Plans.FindByID(ctx, planID)
	if err != nil {
		return options
	}
	return append(options, view.PlanOption{ID: plan.ID, Name: plan.Name})
}

func parseSubscriptionStatus(value string) (domain.SubscriptionStatus, error) {
	status := domain.SubscriptionStatus(strings.ToLower(value))
	if status == "" {
		return domain.SubscriptionActive, nil
	}
	if !status.IsValid() {
		return "", errors.New("Status invalido.")
	}
	return status, nil
}

func normalizeSubscriptionStatus(value string) string {
	switch strings.ToLower(value) {
	case "", "all":
		return "all"
	case string(domain.SubscriptionEnded):
		return string(domain.SubscriptionEnded)
	case string(domain.SubscriptionCanceled):
		return string(domain.SubscriptionCanceled)
	case string(domain.SubscriptionSuspended):
		return string(domain.SubscriptionSuspended)
	default:
		return string(domain.SubscriptionActive)
	}
}

func subscriptionStatusPresentation(status domain.SubscriptionStatus) (string, string) {
	switch status {
	case domain.SubscriptionEnded:
		return "Encerrada", "rounded-full bg-slate-700/50 px-3 py-1 text-slate-300"
	case domain.SubscriptionCanceled:
		return "Cancelada", "rounded-full bg-rose-400/10 px-3 py-1 text-rose-200"
	case domain.SubscriptionSuspended:
		return "Suspensa", "rounded-full bg-amber-400/10 px-3 py-1 text-amber-200"
	default:
		return "Ativa", "rounded-full bg-emerald-400/10 px-3 py-1 text-emerald-200"
	}
}

func parsePriceCentsOptional(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	return parsePriceCents(value)
}

func formatDateBRValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("02/01/2006")
}

func parsePaymentDay(value string) (int, error) {
	if value == "" {
		return 0, errors.New("Dia do pagamento e obrigatorio.")
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 31 {
		return 0, errors.New("Dia do pagamento invalido. Use um numero de 1 a 31.")
	}
	return parsed, nil
}

func formatInt(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}
