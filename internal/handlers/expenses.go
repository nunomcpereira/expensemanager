package handlers

import (
	"bytes"
	"expensemanager/internal/database"
	"expensemanager/internal/i18n"
	"expensemanager/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	db   *database.DB
	tmpl *template.Template
	i18n *i18n.Manager
}

func NewHandler(db *database.DB, tmpl *template.Template) *Handler {
	return &Handler{
		db:   db,
		tmpl: tmpl,
	}
}

// UpdateI18n sets the i18n manager for the handler
func (h *Handler) UpdateI18n(manager *i18n.Manager) {
	h.i18n = manager
}

// TemplateData holds data to be passed to templates
type TemplateData struct {
	CurrentMonth       time.Time
	PreviousMonth      time.Time
	NextMonth          time.Time
	Expenses           []models.Expense
	MonthTotal         float64
	DailyAverage       float64
	Categories         []string
	CategoryTotals     map[string]float64
	Lang               string
	AvailableLanguages []string
	Error              string
	Success            string
	MonthProgress      float64
	DailyTrend         float64
	// Analytics fields
	TotalSpent     float64
	MonthlyTotals  []models.MonthlyTotal
	MonthlyAverage float64
}

// GetTemplateData prepares common template data
func (h *Handler) GetTemplateData(r *http.Request) *TemplateData {
	lang := i18n.GetLang(r.Context())
	return &TemplateData{
		Lang:               lang,
		AvailableLanguages: h.i18n.GetAvailableLanguages(),
		Categories:         models.Categories(),
	}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Get base template data
	data := h.GetTemplateData(r)

	// Get current month
	currentMonth := time.Now()
	data.CurrentMonth = currentMonth
	data.PreviousMonth = currentMonth.AddDate(0, -1, 0)
	data.NextMonth = currentMonth.AddDate(0, 1, 0)

	// Calculate month progress
	daysInMonth := float64(time.Date(currentMonth.Year(), currentMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())
	daysPassed := float64(currentMonth.Day())
	data.MonthProgress = (daysPassed / daysInMonth) * 100

	// Get expenses for current month
	year, month := currentMonth.Year(), int(currentMonth.Month())
	expenses, err := h.db.GetExpensesByMonth(year, month)
	if err != nil {
		http.Error(w, "Failed to retrieve expenses", http.StatusInternalServerError)
		log.Println("Error fetching expenses:", err)
		return
	}
	data.Expenses = expenses

	// Calculate summary statistics
	categoryTotals := make(map[string]float64)
	var total float64

	for _, exp := range expenses {
		total += exp.Amount
		categoryTotals[exp.Category] += exp.Amount
	}
	data.CategoryTotals = categoryTotals
	data.MonthTotal = total

	// Calculate daily average
	if len(expenses) > 0 {
		data.DailyAverage = total / daysInMonth
	}

	// Buffer the template output before writing to ResponseWriter
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, "index.html", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
		return
	}

	// Write buffered output to ResponseWriter
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

func (h *Handler) HandleExpenses(w http.ResponseWriter, r *http.Request) {
	// Get base template data
	data := h.GetTemplateData(r)

	selectedMonth := r.URL.Query().Get("selected-month")
	if selectedMonth == "" {
		selectedMonth = time.Now().Format("2006-01")
	}

	// Parse the selected month
	monthDate, err := time.Parse("2006-01", selectedMonth)
	if err != nil {
		http.Error(w, "Invalid month format", http.StatusBadRequest)
		return
	}

	year, month := monthDate.Year(), int(monthDate.Month())
	expenses, err := h.db.GetExpensesByMonth(year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Expenses = expenses

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "expenses-table", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleAddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get base template data
	data := h.GetTemplateData(r)

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	category := r.FormValue("category")
	dateStr := r.FormValue("date")

	// Parse the date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Create expense model
	expense := &models.Expense{
		Amount:      amount,
		Description: description,
		Category:    category,
		Date:        date,
	}

	if err := h.db.AddExpense(expense); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the month from the expense date
	year, month := date.Year(), int(date.Month())
	expenses, err := h.db.GetExpensesByMonth(year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Expenses = expenses

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "updateSummary")
	if err := h.tmpl.ExecuteTemplate(w, "expenses-table", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get base template data
	data := h.GetTemplateData(r)

	// Parse the ID as int64
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteExpense(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the selected month from the query parameters
	selectedMonth := r.URL.Query().Get("selected-month")
	if selectedMonth == "" {
		selectedMonth = time.Now().Format("2006-01")
	}

	// Parse the selected month
	monthDate, err := time.Parse("2006-01", selectedMonth)
	if err != nil {
		http.Error(w, "Invalid month format", http.StatusBadRequest)
		return
	}

	year, month := monthDate.Year(), int(monthDate.Month())
	expenses, err := h.db.GetExpensesByMonth(year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Expenses = expenses

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "updateSummary")
	if err := h.tmpl.ExecuteTemplate(w, "expenses-table", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	// Get base template data
	data := h.GetTemplateData(r)

	selectedMonth := r.URL.Query().Get("selected-month")
	if selectedMonth == "" {
		selectedMonth = time.Now().Format("2006-01")
	}

	// Parse the month to calculate days and progress
	monthDate, _ := time.Parse("2006-01", selectedMonth)
	daysInMonth := float64(time.Date(monthDate.Year(), monthDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())

	// If it's the current month, calculate actual progress, otherwise 100%
	now := time.Now()
	if monthDate.Year() == now.Year() && monthDate.Month() == now.Month() {
		daysPassed := float64(now.Day())
		data.MonthProgress = (daysPassed / daysInMonth) * 100
	} else if monthDate.Before(now) {
		data.MonthProgress = 100
	} else {
		data.MonthProgress = 0
	}

	// Get expenses for the selected month
	expenses, err := h.db.GetExpensesByMonth(monthDate.Year(), int(monthDate.Month()))
	if err != nil {
		http.Error(w, "Failed to get expenses", http.StatusInternalServerError)
		return
	}
	data.Expenses = expenses

	// Calculate summary statistics
	categoryTotals := make(map[string]float64)
	var total float64
	var todayTotal float64
	var yesterdayTotal float64
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	for _, exp := range expenses {
		total += exp.Amount
		categoryTotals[exp.Category] += exp.Amount

		// Calculate today's and yesterday's totals
		expDate := exp.Date.Format("2006-01-02")
		if expDate == today {
			todayTotal += exp.Amount
		} else if expDate == yesterday {
			yesterdayTotal += exp.Amount
		}
	}
	data.CategoryTotals = categoryTotals
	data.MonthTotal = total

	// Calculate daily average
	if len(expenses) > 0 {
		data.DailyAverage = total / daysInMonth
	}

	// Calculate daily trend
	if yesterdayTotal > 0 {
		data.DailyTrend = ((todayTotal - yesterdayTotal) / yesterdayTotal) * 100
	} else if todayTotal > 0 {
		data.DailyTrend = 100 // 100% increase from 0
	} else {
		data.DailyTrend = 0 // No change
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "summary-cards", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
