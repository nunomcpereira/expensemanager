package database

import (
	"database/sql"
	"strings"

	"expensemanager/internal/models"
)

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Initialize() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount REAL NOT NULL,
			description TEXT NOT NULL,
			category TEXT NOT NULL,
			date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (db *DB) GetExpenses() ([]models.Expense, error) {
	rows, err := db.Query(`
		SELECT id, amount, description, category, date, created_at 
		FROM expenses 
		ORDER BY date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Category, &e.Date, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (db *DB) GetAnalytics() (models.Analytics, error) {
	analytics := models.Analytics{
		CategoryTotals: make(map[string]float64),
		MonthlyTotals:  make([]models.MonthlyTotal, 0),
		TotalSpent:     0.0,
		MonthlyAverage: 0.0,
	}

	// Get total spent
	err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses").Scan(&analytics.TotalSpent)
	if err != nil {
		return analytics, err
	}

	// Get category totals
	rows, err := db.Query(`
		SELECT category, COALESCE(SUM(amount), 0) as total
		FROM expenses
		GROUP BY category
		ORDER BY total DESC
	`)
	if err != nil {
		return analytics, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			return analytics, err
		}
		analytics.CategoryTotals[category] = total
	}

	// Get monthly totals for the last 12 months
	rows, err = db.Query(`
		SELECT 
			strftime('%Y-%m', date) as month,
			COALESCE(SUM(amount), 0) as total
		FROM expenses
		WHERE date >= date('now', '-12 months')
		GROUP BY month
		ORDER BY month ASC
	`)
	if err != nil {
		return analytics, err
	}
	defer rows.Close()

	for rows.Next() {
		var monthTotal models.MonthlyTotal
		if err := rows.Scan(&monthTotal.Month, &monthTotal.Total); err != nil {
			return analytics, err
		}
		analytics.MonthlyTotals = append(analytics.MonthlyTotals, monthTotal)
	}

	// Calculate monthly average
	if len(analytics.MonthlyTotals) > 0 {
		analytics.MonthlyAverage = analytics.TotalSpent / float64(len(analytics.MonthlyTotals))
	}

	return analytics, nil
}

func (db *DB) AddExpense(amount float64, description, category, date string) error {
	// Convert category to lowercase
	category = strings.ToLower(category)

	_, err := db.Exec(`
		INSERT INTO expenses (amount, description, category, date)
		VALUES (?, ?, ?, ?)
	`, amount, description, category, date)
	return err
}

func (db *DB) DeleteExpense(id string) error {
	_, err := db.Exec("DELETE FROM expenses WHERE id = ?", id)
	return err
}

func (db *DB) ClearExpenses() error {
	_, err := db.Exec("DELETE FROM expenses")
	return err
}

func (db *DB) GetExpensesForMonth(month string) ([]models.Expense, error) {
	query := `
		SELECT id, amount, description, category, date, created_at 
		FROM expenses 
		WHERE strftime('%Y-%m', date) = ?
		ORDER BY date DESC
	`
	rows, err := db.Query(query, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Category, &e.Date, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}
