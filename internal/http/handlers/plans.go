package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) PlansIndex(w http.ResponseWriter, r *http.Request) {
	data := h.buildPlansData(r)

	h.renderPage(w, r, page("Planos", view.PlansPage(data)))
}

func (h *Handler) PlansList(w http.ResponseWriter, r *http.Request) {
	data := h.buildPlansData(r)
	h.renderComponent(w, r, view.PlansList(data))
}

func (h *Handler) PlansNew(w http.ResponseWriter, r *http.Request) {
	data := planFormCreateData()
	h.renderPage(w, r, page("Novo plano", view.PlanFormPage(data)))
}

func (h *Handler) PlansCreate(w http.ResponseWriter, r *http.Request) {
	data := planFormCreateData()
	plan, err := parsePlanForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		h.renderFormError(w, r, data.Title, view.PlanFormPage(data))
		return
	}

	if h.services.Plans == nil {
		data.Error = "Servico de planos indisponivel."
		h.renderFormError(w, r, data.Title, view.PlanFormPage(data))
		return
	}

	_, err = h.services.Plans.Create(r.Context(), plan)
	if err != nil {
		data.Error = "Nao foi possivel salvar o plano."
		h.renderFormError(w, r, data.Title, view.PlanFormPage(data))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/plans")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/plans", http.StatusSeeOther)
}

func (h *Handler) PlansEdit(w http.ResponseWriter, r *http.Request) {
	planID := chi.URLParam(r, "planID")
	if h.services.Plans == nil {
		http.NotFound(w, r)
		return
	}

	plan, err := h.services.Plans.FindByID(r.Context(), planID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		observability.Logger(r.Context()).Error("failed to load plan", "err", err)
		http.Error(w, "Erro ao carregar plano.", http.StatusInternalServerError)
		return
	}

	data := planFormEditData(plan)
	h.renderPage(w, r, page(data.Title, view.PlanFormPage(data)))
}

func (h *Handler) PlansUpdate(w http.ResponseWriter, r *http.Request) {
	planID := chi.URLParam(r, "planID")
	data := planFormEditData(domain.Plan{ID: planID})
	plan, err := parsePlanForm(r, &data)
	if err != nil {
		data.Error = err.Error()
		h.renderFormError(w, r, data.Title, view.PlanFormPage(data))
		return
	}

	if h.services.Plans == nil {
		data.Error = "Servico de planos indisponivel."
		h.renderFormError(w, r, data.Title, view.PlanFormPage(data))
		return
	}

	plan.ID = planID
	_, err = h.services.Plans.Update(r.Context(), plan)
	if err != nil {
		data.Error = "Nao foi possivel atualizar o plano."
		h.renderFormError(w, r, data.Title, view.PlanFormPage(data))
		return
	}

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/plans")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, "/plans", http.StatusSeeOther)
}

func (h *Handler) PlansDelete(w http.ResponseWriter, r *http.Request) {
	planID := chi.URLParam(r, "planID")
	if h.services.Plans == nil {
		http.NotFound(w, r)
		return
	}

	_, err := h.services.Plans.Deactivate(r.Context(), planID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		observability.Logger(r.Context()).Error("failed to deactivate plan", "err", err)
		http.Error(w, "Erro ao excluir plano.", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		data := h.buildPlansData(r)
		h.renderComponent(w, r, view.PlansList(data))
		return
	}

	http.Redirect(w, r, "/plans", http.StatusSeeOther)
}

func formatCents(cents int64) string {
	return formatBRL(cents)
}

func formatCentsInput(cents int64) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

func formatBRL(cents int64) string {
	value := fmt.Sprintf("%.2f", float64(cents)/100)
	value = strings.ReplaceAll(value, ".", ",")
	return "R$ " + value
}

func planFormCreateData() view.PlanFormData {
	return view.PlanFormData{
		Title:       "Novo plano",
		Action:      "/plans",
		SubmitLabel: "Criar plano",
		Active:      true,
	}
}

func planFormEditData(plan domain.Plan) view.PlanFormData {
	return view.PlanFormData{
		Title:        "Editar plano",
		Action:       "/plans/" + plan.ID,
		SubmitLabel:  "Salvar",
		DeleteAction: "/plans/" + plan.ID + "/delete",
		ShowDelete:   plan.ID != "",
		Name:         plan.Name,
		DurationDays: strconv.Itoa(plan.DurationDays),
		Price:        formatCentsInput(plan.PriceCents),
		Description:  plan.Description,
		Active:       plan.Active,
	}
}

func parsePlanForm(r *http.Request, data *view.PlanFormData) (domain.Plan, error) {
	if err := r.ParseForm(); err != nil {
		return domain.Plan{}, errors.New("Nao foi possivel ler o formulario.")
	}

	name := strings.TrimSpace(r.FormValue("name"))
	data.Name = name
	if name == "" {
		return domain.Plan{}, errors.New("Nome do plano e obrigatorio.")
	}

	durationRaw := strings.TrimSpace(r.FormValue("duration_days"))
	data.DurationDays = durationRaw
	duration, err := strconv.Atoi(durationRaw)
	if err != nil || duration <= 0 {
		return domain.Plan{}, errors.New("Duracao invalida.")
	}

	priceRaw := strings.TrimSpace(r.FormValue("price"))
	data.Price = priceRaw
	priceCents, err := parsePriceCents(priceRaw)
	if err != nil || priceCents < 0 {
		return domain.Plan{}, errors.New("Preco invalido.")
	}

	description := strings.TrimSpace(r.FormValue("description"))
	data.Description = description

	active := r.FormValue("active") != ""
	data.Active = active

	return domain.Plan{
		Name:         name,
		DurationDays: duration,
		PriceCents:   priceCents,
		Description:  description,
		Active:       active,
	}, nil
}

func parsePriceCents(value string) (int64, error) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return 0, errors.New("preco vazio")
	}

	clean = strings.ReplaceAll(clean, ",", ".")
	parsed, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, err
	}

	return int64(parsed*100 + 0.5), nil
}

func (h *Handler) buildPlansData(r *http.Request) view.PlansPageData {
	data := view.PlansPageData{}

	if h.services.Plans != nil {
		plans, err := h.services.Plans.ListActive(r.Context())
		if err != nil {
			observability.Logger(r.Context()).Error("failed to list plans", "err", err)
		} else {
			data.Items = make([]view.PlanItem, 0, len(plans))
			for _, plan := range plans {
				item := view.PlanItem{
					ID:           plan.ID,
					Name:         plan.Name,
					DurationDays: plan.DurationDays,
					Price:        formatCents(plan.PriceCents),
					Description:  plan.Description,
				}
				data.Items = append(data.Items, item)
			}
		}
	}

	return data
}
