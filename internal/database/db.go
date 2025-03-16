package database

import (
	"database/sql"
	"time"

	"expensemanager/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
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
			id SERIAL PRIMARY KEY,
			amount DECIMAL(10,2) NOT NULL,
			description TEXT NOT NULL,
			category TEXT NOT NULL,
			date DATE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (db *DB) GetExpenses() ([]models.Expense, error) {
	rows, err := db.Query(`
		SELECT id, amount, description, category, date, created_at, updated_at
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
		err := rows.Scan(
			&e.ID,
			&e.Amount,
			&e.Description,
			&e.Category,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (db *DB) GetExpensesByMonth(year int, month int) ([]models.Expense, error) {
	rows, err := db.Query(`
		SELECT id, amount, description, category, date, created_at, updated_at
		FROM expenses 
		WHERE EXTRACT(YEAR FROM date) = $1 
		AND EXTRACT(MONTH FROM date) = $2
		ORDER BY date DESC
	`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(
			&e.ID,
			&e.Amount,
			&e.Description,
			&e.Category,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (db *DB) AddExpense(e *models.Expense) error {
	now := time.Now()
	err := db.QueryRow(`
		INSERT INTO expenses (amount, description, category, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, e.Amount, e.Description, e.Category, e.Date, now, now).Scan(&e.ID)

	if err != nil {
		return err
	}

	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}

func (db *DB) DeleteExpense(id int64) error {
	result, err := db.Exec("DELETE FROM expenses WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
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
			TO_CHAR(date, 'YYYY-MM') as month,
			COALESCE(SUM(amount), 0) as total
		FROM expenses
		WHERE date >= NOW() - INTERVAL '1 year'
		GROUP BY month
		ORDER BY month DESC
	`)
	if err != nil {
		return analytics, err
	}
	defer rows.Close()

	for rows.Next() {
		var mt models.MonthlyTotal
		if err := rows.Scan(&mt.Month, &mt.Total); err != nil {
			return analytics, err
		}
		analytics.MonthlyTotals = append(analytics.MonthlyTotals, mt)
	}

	// Calculate monthly average
	if len(analytics.MonthlyTotals) > 0 {
		analytics.MonthlyAverage = analytics.TotalSpent / float64(len(analytics.MonthlyTotals))
	}

	return analytics, nil
}

func (db *DB) ClearExpenses() error {
	_, err := db.Exec("DELETE FROM expenses")
	return err
}
