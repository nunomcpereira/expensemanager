package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleReports(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, _ := GetUserIDFromContext(r.Context())

	// Get base template data
	data := h.GetTemplateData(r)

	// Get analytics data
	analytics, err := h.db.GetAnalytics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Combine analytics with template data
	data.TotalSpent = analytics.TotalSpent
	data.CategoryTotals = analytics.CategoryTotals
	data.MonthlyTotals = analytics.MonthlyTotals
	data.MonthlyAverage = analytics.MonthlyAverage

	if err := h.tmpl.ExecuteTemplate(w, "reports", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleMonthlyTotals(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, _ := GetUserIDFromContext(r.Context())

	analytics, err := h.db.GetAnalytics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics.MonthlyTotals)
}

func (h *Handler) HandleCategoryTotals(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, _ := GetUserIDFromContext(r.Context())

	analytics, err := h.db.GetAnalytics(userID)
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
