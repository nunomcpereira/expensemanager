package database

import (
	"database/sql"
	"errors"
	"time"

	"expensemanager/internal/models"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Drop existing expenses table if it exists
	_, err = db.Exec(`DROP TABLE IF EXISTS expenses`)
	if err != nil {
		return err
	}

	// Create expenses table with user_id foreign key
	_, err = db.Exec(`
		CREATE TABLE expenses (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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

// User-related functions
func (db *DB) CreateUser(user *models.User, password string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	err = db.QueryRow(`
		INSERT INTO users (email, password, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id
	`, user.Email, string(hashedPassword), user.Name, now).Scan(&user.ID)

	if err != nil {
		return err
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(`
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) AuthenticateUser(email, password string) (*models.User, error) {
	user, err := db.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// Update expense-related functions to include user_id
func (db *DB) GetExpenses(userID int64) ([]models.Expense, error) {
	rows, err := db.Query(`
		SELECT id, user_id, amount, description, category, date, created_at, updated_at
		FROM expenses 
		WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(
			&e.ID,
			&e.UserID,
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

func (db *DB) GetExpensesByMonth(userID int64, year int, month int) ([]models.Expense, error) {
	rows, err := db.Query(`
		SELECT id, user_id, amount, description, category, date, created_at, updated_at
		FROM expenses 
		WHERE user_id = $1
		AND EXTRACT(YEAR FROM date) = $2 
		AND EXTRACT(MONTH FROM date) = $3
		ORDER BY date DESC
	`, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(
			&e.ID,
			&e.UserID,
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
		INSERT INTO expenses (user_id, amount, description, category, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id
	`, e.UserID, e.Amount, e.Description, e.Category, e.Date, now).Scan(&e.ID)

	if err != nil {
		return err
	}

	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}

func (db *DB) DeleteExpense(userID, expenseID int64) error {
	result, err := db.Exec("DELETE FROM expenses WHERE id = $1 AND user_id = $2", expenseID, userID)
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

func (db *DB) GetAnalytics(userID int64) (models.Analytics, error) {
	analytics := models.Analytics{
		CategoryTotals: make(map[string]float64),
		MonthlyTotals:  make([]models.MonthlyTotal, 0),
		TotalSpent:     0.0,
		MonthlyAverage: 0.0,
	}

	// Get total spent
	err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1", userID).Scan(&analytics.TotalSpent)
	if err != nil {
		return analytics, err
	}

	// Get category totals
	rows, err := db.Query(`
		SELECT category, COALESCE(SUM(amount), 0) as total
		FROM expenses
		WHERE user_id = $1
		GROUP BY category
		ORDER BY total DESC
	`, userID)
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
		WHERE user_id = $1
		AND date >= NOW() - INTERVAL '1 year'
		GROUP BY month
		ORDER BY month DESC
	`, userID)
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

func (db *DB) ClearExpenses(userID int64) error {
	_, err := db.Exec("DELETE FROM expenses WHERE user_id = $1", userID)
	return err
}
