package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PabloPavan/jaiu/internal/view"
)

func (h *Handler) PlansIndex(w http.ResponseWriter, r *http.Request) {
	data := view.PlansPageData{}

	if h.plans != nil {
		plans, err := h.plans.ListActive(r.Context())
		if err != nil {
			log.Printf("list plans: %v", err)
		} else {
			data.Items = make([]view.PlanItem, 0, len(plans))
			for _, plan := range plans {
				item := view.PlanItem{
					Name:         plan.Name,
					DurationDays: plan.DurationDays,
					Price:        formatCents(plan.PriceCents),
					Description:  plan.Description,
				}
				data.Items = append(data.Items, item)
			}
		}
	}

	h.renderPage(w, r, page("Planos", view.PlansPage(data)))
}

func formatCents(cents int64) string {
	return fmt.Sprintf("R$ %.2f", float64(cents)/100)
}
