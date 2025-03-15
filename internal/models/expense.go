package models

import "time"

// Expense represents a single expense entry
type Expense struct {
	ID          int64
	Amount      float64
	Category    string
	Description string
	Date        time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Categories returns a list of valid expense categories
func Categories() []string {
	return []string{
		"food",
		"transportation",
		"entertainment",
		"shopping",
		"bills",
		"health",
		"other",
	}
}

type Analytics struct {
	TotalSpent     float64
	CategoryTotals map[string]float64
	MonthlyTotals  []MonthlyTotal
	MonthlyAverage float64
}

type MonthlyTotal struct {
	Month string  `json:"month"`
	Total float64 `json:"total"`
}

type ExpenseJSON struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Date        string  `json:"date"`
	CreatedAt   string  `json:"created_at"`
}
