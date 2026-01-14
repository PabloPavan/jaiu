package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) PaymentsIndex(w http.ResponseWriter, r *http.Request) {
	data := h.buildPaymentsData(r)
	if isHTMX(r) {
		h.renderComponent(w, r, view.PaymentsList(data))
		return
	}
	h.renderPage(w, r, page("Pagamentos", view.PaymentsPage(data)))
}

func (h *Handler) PaymentsNew(w http.ResponseWriter, r *http.Request) {
	data := h.paymentFormCreateData(r)
	h.renderPage(w, r, page("Novo pagamento", view.PaymentFormPage(data)))
}

func (h *Handler) PaymentsCreate(w http.ResponseWriter, r *http.Request) {
	data := h.paymentFormCreateData(r)
	payment, err := h.parsePaymentForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		if isHTMX(r) {
			h.renderComponent(w, r, view.PaymentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
		return
	}

	if h.services.Payments == nil {
		data.Error = "Servico de pagamentos indisponivel."
		if isHTMX(r) {
			h.renderComponent(w, r, view.PaymentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
		return
	}

	if _, err := h.services.Payments.Register(r.Context(), payment); err != nil {
		data.Error = "Nao foi possivel salvar o pagamento."
		if isHTMX(r) {
			h.renderComponent(w, r, view.PaymentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/payments")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/payments", http.StatusSeeOther)
}

func (h *Handler) PaymentsEdit(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	if h.services.Payments == nil {
		http.NotFound(w, r)
		return
	}

	payment, err := h.services.Payments.FindByID(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("load payment: %v", err)
		http.Error(w, "Erro ao carregar pagamento.", http.StatusInternalServerError)
		return
	}

	data := h.paymentFormEditData(r, payment)
	h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
}

func (h *Handler) PaymentsUpdate(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	data := h.paymentFormEditData(r, domain.Payment{ID: paymentID})
	payment, err := h.parsePaymentForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		if isHTMX(r) {
			h.renderComponent(w, r, view.PaymentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
		return
	}

	if h.services.Payments == nil {
		data.Error = "Servico de pagamentos indisponivel."
		if isHTMX(r) {
			h.renderComponent(w, r, view.PaymentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
		return
	}

	payment.ID = paymentID
	if _, err := h.services.Payments.Update(r.Context(), payment); err != nil {
		data.Error = "Nao foi possivel atualizar o pagamento."
		if isHTMX(r) {
			h.renderComponent(w, r, view.PaymentFormPage(data))
			return
		}
		h.renderPage(w, r, page(data.Title, view.PaymentFormPage(data)))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/payments")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/payments", http.StatusSeeOther)
}

func (h *Handler) PaymentsReverse(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	if h.services.Payments == nil {
		http.NotFound(w, r)
		return
	}

	if _, err := h.services.Payments.Reverse(r.Context(), paymentID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("reverse payment: %v", err)
		http.Error(w, "Erro ao estornar pagamento.", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		data := h.buildPaymentsData(r)
		h.renderComponent(w, r, view.PaymentsList(data))
		return
	}

	http.Redirect(w, r, "/payments", http.StatusSeeOther)
}

func (h *Handler) buildPaymentsData(r *http.Request) view.PaymentsPageData {
	subscriptionID := strings.TrimSpace(r.FormValue("subscription_id"))
	status := normalizePaymentStatus(strings.TrimSpace(r.FormValue("status")))

	subscriptions := h.loadSubscriptions(r, nil)
	data := view.PaymentsPageData{
		SubscriptionID: subscriptionID,
		Status:         status,
		Subscriptions:  toSubscriptionOptions(subscriptions, r, h),
	}

	if h.services.Payments == nil {
		return data
	}

	var payments []domain.Payment
	if subscriptionID != "" {
		list, err := h.services.Payments.ListBySubscription(r.Context(), subscriptionID)
		if err != nil {
			log.Printf("list payments: %v", err)
			return data
		}
		payments = list
	} else {
		payments = make([]domain.Payment, 0)
		for _, subscription := range subscriptions {
			list, err := h.services.Payments.ListBySubscription(r.Context(), subscription.ID)
			if err != nil {
				log.Printf("list payments: %v", err)
				continue
			}
			payments = append(payments, list...)
		}
	}

	if len(payments) == 0 {
		return data
	}

	sort.Slice(payments, func(i, j int) bool {
		return payments[i].PaidAt.After(payments[j].PaidAt)
	})

	subscriptionMap := make(map[string]string, len(data.Subscriptions))
	for _, option := range data.Subscriptions {
		subscriptionMap[option.ID] = option.Label
	}

	data.Items = make([]view.PaymentItem, 0, len(payments))
	for _, payment := range payments {
		if status != "all" && string(payment.Status) != status {
			continue
		}

		label := subscriptionMap[payment.SubscriptionID]
		if label == "" {
			label = payment.SubscriptionID
		}

		statusLabel, statusClass := paymentStatusPresentation(payment.Status)
		item := view.PaymentItem{
			ID:                payment.ID,
			SubscriptionLabel: label,
			PaidAt:            formatDateBRValue(payment.PaidAt),
			Amount:            formatBRL(payment.AmountCents),
			MethodLabel:       paymentMethodLabel(payment.Method),
			Reference:         payment.Reference,
			Notes:             payment.Notes,
			Status:            string(payment.Status),
			StatusLabel:       statusLabel,
			StatusClass:       statusClass,
		}
		data.Items = append(data.Items, item)
	}

	return data
}

func (h *Handler) paymentFormCreateData(r *http.Request) view.PaymentFormData {
	now := time.Now()
	data := view.PaymentFormData{
		Title:       "Novo pagamento",
		Action:      "/payments",
		SubmitLabel: "Registrar pagamento",
		PaidAt:      formatDateBRValue(now),
		Status:      string(domain.PaymentConfirmed),
	}
	h.fillPaymentFormOptions(r, &data)
	return data
}

func (h *Handler) paymentFormEditData(r *http.Request, payment domain.Payment) view.PaymentFormData {
	data := view.PaymentFormData{
		Title:          "Editar pagamento",
		Action:         "/payments/" + payment.ID,
		SubmitLabel:    "Salvar",
		DeleteAction:   "/payments/" + payment.ID + "/reverse",
		ShowDelete:     payment.ID != "",
		SubscriptionID: payment.SubscriptionID,
		PaidAt:         formatDateBRValue(payment.PaidAt),
		Amount:         formatAmountInput(payment.AmountCents),
		Method:         string(payment.Method),
		Reference:      payment.Reference,
		Notes:          payment.Notes,
		Status:         string(payment.Status),
	}
	h.fillPaymentFormOptions(r, &data)
	return data
}

func (h *Handler) fillPaymentFormOptions(r *http.Request, data *view.PaymentFormData) {
	subscriptions := h.loadSubscriptions(r, []domain.StudentStatus{domain.StudentActive})
	data.Subscriptions = toSubscriptionOptions(subscriptions, r, h)
	if h.services.Subscriptions == nil || h.services.Students == nil || h.services.Plans == nil {
		data.Error = "Dependencias de pagamentos indisponiveis."
	}

	data.Subscriptions = ensureSubscriptionOption(r.Context(), data.Subscriptions, data.SubscriptionID, h)
}

func (h *Handler) parsePaymentForm(r *http.Request, data *view.PaymentFormData) (domain.Payment, error) {
	if err := r.ParseForm(); err != nil {
		return domain.Payment{}, errors.New("Nao foi possivel ler o formulario.")
	}

	subscriptionID := strings.TrimSpace(r.FormValue("subscription_id"))
	data.SubscriptionID = subscriptionID
	if subscriptionID == "" {
		return domain.Payment{}, errors.New("Assinatura e obrigatoria.")
	}

	paidRaw := strings.TrimSpace(r.FormValue("paid_at"))
	data.PaidAt = paidRaw
	paidAt, err := parseDateInput(paidRaw)
	if err != nil || paidAt == nil {
		return domain.Payment{}, errors.New("Data do pagamento invalida. Use o formato dd/mm/aaaa.")
	}

	amountRaw := strings.TrimSpace(r.FormValue("amount"))
	data.Amount = amountRaw
	amountCents, err := parsePriceCents(amountRaw)
	if err != nil || amountCents <= 0 {
		return domain.Payment{}, errors.New("Valor invalido.")
	}

	methodValue := strings.TrimSpace(r.FormValue("method"))
	method := parsePaymentMethod(methodValue)
	data.Method = string(method)

	statusValue := strings.TrimSpace(r.FormValue("status"))
	status, err := parsePaymentStatus(statusValue)
	if err != nil {
		return domain.Payment{}, err
	}
	data.Status = string(status)

	reference := strings.TrimSpace(r.FormValue("reference"))
	notes := strings.TrimSpace(r.FormValue("notes"))
	data.Reference = reference
	data.Notes = notes

	return domain.Payment{
		SubscriptionID: subscriptionID,
		PaidAt:         *paidAt,
		AmountCents:    amountCents,
		Method:         method,
		Reference:      reference,
		Notes:          notes,
		Status:         status,
	}, nil
}

func (h *Handler) loadSubscriptions(r *http.Request, statuses []domain.StudentStatus) []domain.Subscription {
	if h.services.Subscriptions == nil {
		return nil
	}

	students := h.loadStudents(r, statuses, 500)
	subscriptions := make([]domain.Subscription, 0)
	for _, student := range students {
		list, err := h.services.Subscriptions.ListByStudent(r.Context(), student.ID)
		if err != nil {
			log.Printf("list subscriptions: %v", err)
			continue
		}
		subscriptions = append(subscriptions, list...)
	}
	return subscriptions
}

func toSubscriptionOptions(subscriptions []domain.Subscription, r *http.Request, h *Handler) []view.SubscriptionOption {
	if len(subscriptions) == 0 {
		return nil
	}

	planOptions := h.listPlanOptions(r)
	planMap := make(map[string]string, len(planOptions))
	for _, option := range planOptions {
		planMap[option.ID] = option.Name
	}

	studentOptions := h.listStudentOptions(r)
	studentMap := make(map[string]string, len(studentOptions))
	for _, option := range studentOptions {
		studentMap[option.ID] = option.Name
	}

	options := make([]view.SubscriptionOption, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		studentName := studentMap[subscription.StudentID]
		planName := planMap[subscription.PlanID]
		label := strings.TrimSpace(studentName + " · " + planName)
		options = append(options, view.SubscriptionOption{
			ID:    subscription.ID,
			Label: label,
		})
	}
	return options
}

func ensureSubscriptionOption(ctx context.Context, options []view.SubscriptionOption, subscriptionID string, h *Handler) []view.SubscriptionOption {
	if subscriptionID == "" || h.services.Subscriptions == nil {
		return options
	}
	for _, option := range options {
		if option.ID == subscriptionID {
			return options
		}
	}
	subscription, err := h.services.Subscriptions.FindByID(ctx, subscriptionID)
	if err != nil {
		return options
	}

	studentName := subscription.StudentID
	if h.services.Students != nil {
		if student, err := h.services.Students.FindByID(ctx, subscription.StudentID); err == nil {
			studentName = student.FullName
		}
	}
	planName := subscription.PlanID
	if h.services.Plans != nil {
		if plan, err := h.services.Plans.FindByID(ctx, subscription.PlanID); err == nil {
			planName = plan.Name
		}
	}

	label := strings.TrimSpace(studentName + " · " + planName)
	return append(options, view.SubscriptionOption{
		ID:    subscription.ID,
		Label: label,
	})
}

func parsePaymentMethod(value string) domain.PaymentMethod {
	switch strings.ToLower(value) {
	case string(domain.PaymentPix):
		return domain.PaymentPix
	case string(domain.PaymentCard):
		return domain.PaymentCard
	case string(domain.PaymentTransfer):
		return domain.PaymentTransfer
	case string(domain.PaymentOther):
		return domain.PaymentOther
	default:
		return domain.PaymentCash
	}
}

func parsePaymentStatus(value string) (domain.PaymentStatus, error) {
	switch strings.ToLower(value) {
	case "", string(domain.PaymentConfirmed):
		return domain.PaymentConfirmed, nil
	case string(domain.PaymentReversed):
		return domain.PaymentReversed, nil
	default:
		return "", errors.New("Status invalido.")
	}
}

func normalizePaymentStatus(value string) string {
	switch strings.ToLower(value) {
	case "", "all":
		return "all"
	case string(domain.PaymentReversed):
		return string(domain.PaymentReversed)
	default:
		return string(domain.PaymentConfirmed)
	}
}

func paymentStatusPresentation(status domain.PaymentStatus) (string, string) {
	switch status {
	case domain.PaymentReversed:
		return "Estornado", "rounded-full bg-rose-400/10 px-3 py-1 text-rose-200"
	default:
		return "Confirmado", "rounded-full bg-emerald-400/10 px-3 py-1 text-emerald-200"
	}
}

func paymentMethodLabel(method domain.PaymentMethod) string {
	switch method {
	case domain.PaymentPix:
		return "Pix"
	case domain.PaymentCard:
		return "Cartao"
	case domain.PaymentTransfer:
		return "Transferencia"
	case domain.PaymentOther:
		return "Outro"
	default:
		return "Dinheiro"
	}
}

func formatAmountInput(cents int64) string {
	value := formatCentsInput(cents)
	return strings.ReplaceAll(value, ".", ",")
}
