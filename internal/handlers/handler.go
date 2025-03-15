package handlers

import (
	"html/template"
	"time"

	"expensemanager/internal/i18n"
	"expensemanager/internal/models"
)

// Handler holds dependencies for handling HTTP requests
type Handler struct {
	templates *template.Template
	i18n      *i18n.Manager
}

// NewHandler creates a new Handler instance
func NewHandler(templates *template.Template, i18nManager *i18n.Manager) *Handler {
	return &Handler{
		templates: templates,
		i18n:      i18nManager,
	}
}

// TemplateData holds data to be passed to templates
type TemplateData struct {
	CurrentMonth       time.Time
	PreviousMonth      time.Time
	NextMonth          time.Time
	Expenses           []models.Expense
	MonthlyTotal       float64
	DailyAverage       float64
	Categories         []string
	CategoryTotals     map[string]float64
	Lang               string
	AvailableLanguages []string
	Error              string
	Success            string
}
