package models

import "time"

type Expense struct {
	ID          int64
	Amount      float64
	Description string
	Category    string
	Date        time.Time
	CreatedAt   time.Time
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
