package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"expensemanager/internal/models"
)

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *Handler) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	// Get base template data
	data := h.GetTemplateData(r)

	if err := h.tmpl.ExecuteTemplate(w, "admin", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleClearExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, _ := GetUserIDFromContext(r.Context())

	if err := h.db.ClearExpenses(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expenses, err := h.db.GetExpenses(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "updateSummary")
	h.tmpl.ExecuteTemplate(w, "expenses-table.html", expenses)
}

func (h *Handler) HandleDownloadExpenses(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, _ := GetUserIDFromContext(r.Context())

	expenses, err := h.db.GetExpenses(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expensesJSON := make([]models.ExpenseJSON, len(expenses))
	for i, e := range expenses {
		expensesJSON[i] = models.ExpenseJSON{
			ID:          e.ID,
			UserID:      e.UserID,
			Amount:      e.Amount,
			Description: e.Description,
			Category:    e.Category,
			Date:        e.Date.Format("2006-01-02"),
			CreatedAt:   e.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=expenses.json")
	json.NewEncoder(w).Encode(expensesJSON)
}

func (h *Handler) HandleUploadExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, _ := GetUserIDFromContext(r.Context())

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file contents
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Parse JSON
	var expenses []models.ExpenseJSON
	if err := json.Unmarshal(fileBytes, &expenses); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate and add expenses
	for _, e := range expenses {
		// Parse date
		date, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid date format for expense %d", e.ID), http.StatusBadRequest)
			return
		}

		// Validate category
		validCategory := false
		for _, cat := range models.Categories() {
			if strings.EqualFold(e.Category, cat) {
				validCategory = true
				break
			}
		}
		if !validCategory {
			http.Error(w, fmt.Sprintf("Invalid category for expense %d", e.ID), http.StatusBadRequest)
			return
		}

		// Create expense
		expense := &models.Expense{
			UserID:      userID,
			Amount:      e.Amount,
			Description: e.Description,
			Category:    e.Category,
			Date:        date,
		}

		if err := h.db.AddExpense(expense); err != nil {
			http.Error(w, fmt.Sprintf("Failed to add expense %d", e.ID), http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	response := UploadResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully uploaded %d expenses", len(expenses)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
