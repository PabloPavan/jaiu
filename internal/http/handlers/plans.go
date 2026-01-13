package handlers

import (
	"fmt"
	"log"
	"net/http"
)

type planListItem struct {
	Name         string
	DurationDays int
	Price        string
	Description  string
}

type planListViewData struct {
	Items []planListItem
}

func (h *Handler) PlansIndex(w http.ResponseWriter, r *http.Request) {
	data := planListViewData{}

	if h.plans != nil {
		plans, err := h.plans.ListActive(r.Context())
		if err != nil {
			log.Printf("list plans: %v", err)
		} else {
			data.Items = make([]planListItem, 0, len(plans))
			for _, plan := range plans {
				item := planListItem{
					Name:         plan.Name,
					DurationDays: plan.DurationDays,
					Price:        formatCents(plan.PriceCents),
					Description:  plan.Description,
				}
				data.Items = append(data.Items, item)
			}
		}
	}

	h.renderPage(w, r, page("Planos", "page/plans/index", data))
}

func formatCents(cents int64) string {
	return fmt.Sprintf("R$ %.2f", float64(cents)/100)
}
