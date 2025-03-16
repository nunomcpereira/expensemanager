package models

import "time"

// Expense represents a single expense entry
type Expense struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Categories returns a list of valid expense categories
func Categories() []string {
	return []string{
		"food",
		"transportation",
		"housing",
		"utilities",
		"entertainment",
		"healthcare",
		"shopping",
		"education",
		"travel",
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
	UserID      int64   `json:"user_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Date        string  `json:"date"`
	CreatedAt   string  `json:"created_at"`
}
