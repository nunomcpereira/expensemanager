package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleReports(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.db.GetAnalytics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "reports", analytics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleMonthlyTotals(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.db.GetAnalytics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics.MonthlyTotals)
}

func (h *Handler) HandleCategoryTotals(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.db.GetAnalytics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	categories := make([]map[string]interface{}, 0)
	for category, total := range analytics.CategoryTotals {
		categories = append(categories, map[string]interface{}{
			"category": category,
			"total":    total,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
