package migration

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"expensemanager/internal/models"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Migration handles the migration from SQLite to PostgreSQL
type Migration struct {
	sourceDB *sql.DB
	targetDB *sql.DB
}

// NewMigration creates a new migration instance
func NewMigration(sqliteDBPath, postgresConnStr string) (*Migration, error) {
	sourceDB, err := sql.Open("sqlite3", sqliteDBPath)
	if err != nil {
		return nil, fmt.Errorf("error opening SQLite database: %v", err)
	}

	targetDB, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		sourceDB.Close()
		return nil, fmt.Errorf("error opening PostgreSQL database: %v", err)
	}

	return &Migration{
		sourceDB: sourceDB,
		targetDB: targetDB,
	}, nil
}

// Close closes both database connections
func (m *Migration) Close() {
	if m.sourceDB != nil {
		m.sourceDB.Close()
	}
	if m.targetDB != nil {
		m.targetDB.Close()
	}
}

// CreatePostgresSchema creates the necessary tables in PostgreSQL
func (m *Migration) CreatePostgresSchema() error {
	_, err := m.targetDB.Exec(`
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
	if err != nil {
		return fmt.Errorf("error creating expenses table: %v", err)
	}

	return nil
}

// MigrateData migrates all data from SQLite to PostgreSQL
func (m *Migration) MigrateData() error {
	// Start a transaction
	tx, err := m.targetDB.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO expenses (amount, description, category, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("error preparing insert statement: %v", err)
	}
	defer stmt.Close()

	// Query all expenses from SQLite
	rows, err := m.sourceDB.Query(`
		SELECT amount, description, category, date, created_at
		FROM expenses
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("error querying source database: %v", err)
	}
	defer rows.Close()

	// Migrate each expense
	count := 0
	for rows.Next() {
		var expense models.Expense
		var dateStr, createdAtStr string

		err := rows.Scan(
			&expense.Amount,
			&expense.Description,
			&expense.Category,
			&dateStr,
			&createdAtStr,
		)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		// Parse dates
		expense.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("error parsing date: %v", err)
		}

		expense.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			expense.CreatedAt = time.Now()
		}

		expense.UpdatedAt = time.Now()

		// Insert into PostgreSQL
		_, err = stmt.Exec(
			expense.Amount,
			expense.Description,
			expense.Category,
			expense.Date,
			expense.CreatedAt,
			expense.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error inserting expense: %v", err)
		}

		count++
		if count%100 == 0 {
			log.Printf("Migrated %d expenses...", count)
		}
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	log.Printf("Successfully migrated %d expenses", count)
	return nil
}
