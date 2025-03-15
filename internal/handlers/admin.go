package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	if err := h.db.ClearExpenses(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expenses, err := h.db.GetExpenses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "updateSummary")
	h.tmpl.ExecuteTemplate(w, "expenses-table.html", expenses)
}

func (h *Handler) HandleDownloadExpenses(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.db.GetExpenses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expensesJSON := make([]models.ExpenseJSON, len(expenses))
	for i, e := range expenses {
		expensesJSON[i] = models.ExpenseJSON{
			ID:          e.ID,
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

	// Parse the multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		sendJSONResponse(w, false, "Failed to parse form: "+err.Error())
		return
	}

	// Get the file from the form
	file, _, err := r.FormFile("file")
	if err != nil {
		sendJSONResponse(w, false, "Failed to get file: "+err.Error())
		return
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		sendJSONResponse(w, false, "Failed to read file: "+err.Error())
		return
	}

	// Parse the JSON content
	var uploadData struct {
		Expenses []struct {
			Amount      float64 `json:"amount"`
			Description string  `json:"description"`
			Category    string  `json:"category"`
			Date        string  `json:"date"`
		} `json:"expenses"`
	}

	if err := json.Unmarshal(content, &uploadData); err != nil {
		sendJSONResponse(w, false, "Invalid JSON format: "+err.Error())
		return
	}

	// Validate and insert each expense
	var errors []string
	for i, expense := range uploadData.Expenses {
		// Validate required fields
		if expense.Amount <= 0 {
			errors = append(errors, fmt.Sprintf("Expense %d: Amount must be greater than 0", i+1))
			continue
		}
		if strings.TrimSpace(expense.Description) == "" {
			errors = append(errors, fmt.Sprintf("Expense %d: Description is required", i+1))
			continue
		}
		if strings.TrimSpace(expense.Category) == "" {
			errors = append(errors, fmt.Sprintf("Expense %d: Category is required", i+1))
			continue
		}
		if strings.TrimSpace(expense.Date) == "" {
			errors = append(errors, fmt.Sprintf("Expense %d: Date is required", i+1))
			continue
		}

		// Insert the expense
		if err := h.db.AddExpense(expense.Amount, expense.Description, expense.Category, expense.Date); err != nil {
			errors = append(errors, fmt.Sprintf("Expense %d: Failed to insert: %v", i+1, err))
			continue
		}
	}

	if len(errors) > 0 {
		sendJSONResponse(w, false, "Some expenses failed to upload:\n"+strings.Join(errors, "\n"))
		return
	}

	sendJSONResponse(w, true, "All expenses uploaded successfully")
}

func sendJSONResponse(w http.ResponseWriter, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{
		Success: success,
		Message: message,
	})
}
