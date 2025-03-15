package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"expensemanager/internal/database"
	"expensemanager/internal/models"
)

type Handler struct {
	db   *database.DB
	tmpl *template.Template
}

func NewHandler(db *database.DB, tmpl *template.Template) *Handler {
	return &Handler{db: db, tmpl: tmpl}
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Get current month
	currentMonth := time.Now().Format("2006-01")

	// Get expenses for current month
	expenses, err := h.db.GetExpensesForMonth(currentMonth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate summary statistics
	var total float64
	categories := make(map[string]bool)

	for _, exp := range expenses {
		total += exp.Amount
		categories[exp.Category] = true
	}

	// Calculate days in current month
	now := time.Now()
	daysInMonth := float64(time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())

	// Calculate daily average
	dailyAverage := 0.0
	if len(expenses) > 0 {
		dailyAverage = total / daysInMonth
	}

	// Create summary data
	summaryData := SummaryData{
		MonthTotal:    total,
		DailyAverage:  dailyAverage,
		CategoryCount: len(categories),
	}

	// Create combined data structure
	data := struct {
		Expenses []models.Expense
		SummaryData
	}{
		Expenses:    expenses,
		SummaryData: summaryData,
	}

	if err := h.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleExpenses(w http.ResponseWriter, r *http.Request) {
	var expenses []models.Expense
	var err error

	selectedMonth := r.URL.Query().Get("selected-month")
	if selectedMonth == "" {
		selectedMonth = time.Now().Format("2006-01")
	}

	expenses, err = h.db.GetExpensesForMonth(selectedMonth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.ExecuteTemplate(w, "expenses-table", expenses)
}

func (h *Handler) HandleAddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	category := r.FormValue("category")
	date := r.FormValue("date")

	if err := h.db.AddExpense(amount, description, category, date); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the expense date to get the month we should display
	expenseDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Use the month from the expense date
	monthToShow := expenseDate.Format("2006-01")

	// Get expenses for that month
	expenses, err := h.db.GetExpensesForMonth(monthToShow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "updateSummary")
	h.tmpl.ExecuteTemplate(w, "expenses-table.html", expenses)
}

func (h *Handler) HandleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if err := h.db.DeleteExpense(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the selected month from the query parameters
	selectedMonth := r.URL.Query().Get("selected-month")
	if selectedMonth == "" {
		selectedMonth = time.Now().Format("2006-01")
	}

	// Get expenses for the selected month
	expenses, err := h.db.GetExpensesForMonth(selectedMonth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "updateSummary")
	h.tmpl.ExecuteTemplate(w, "expenses-table.html", expenses)
}

type SummaryData struct {
	MonthTotal    float64
	DailyAverage  float64
	CategoryCount int
}

func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	selectedMonth := r.URL.Query().Get("selected-month")
	if selectedMonth == "" {
		selectedMonth = time.Now().Format("2006-01")
	}

	// Get expenses for the selected month
	expenses, err := h.db.GetExpensesForMonth(selectedMonth)
	if err != nil {
		http.Error(w, "Failed to get expenses", http.StatusInternalServerError)
		return
	}

	// Calculate summary statistics
	var total float64
	categories := make(map[string]bool)

	for _, exp := range expenses {
		total += exp.Amount
		categories[exp.Category] = true
	}

	// Parse the month to calculate days
	monthDate, _ := time.Parse("2006-01", selectedMonth)
	daysInMonth := float64(time.Date(monthDate.Year(), monthDate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())

	// Calculate daily average
	dailyAverage := 0.0
	if len(expenses) > 0 {
		dailyAverage = total / daysInMonth
	}

	data := SummaryData{
		MonthTotal:    total,
		DailyAverage:  dailyAverage,
		CategoryCount: len(categories),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.ExecuteTemplate(w, "summary-cards", data)
}
