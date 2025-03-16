package main

import (
	"flag"
	"log"
	"os"

	"expensemanager/internal/config"
	"expensemanager/internal/database/migration"
)

func main() {
	// Parse command line flags
	sqliteDBPath := flag.String("sqlite", "db/expenses.db", "Path to SQLite database file")
	flag.Parse()

	// Check if SQLite database exists
	if _, err := os.Stat(*sqliteDBPath); os.IsNotExist(err) {
		log.Fatalf("SQLite database file not found: %s", *sqliteDBPath)
	}

	// Get PostgreSQL configuration
	dbConfig := config.NewDBConfig()
	postgresConnStr := dbConfig.PostgresConnectionString()

	// Create migration instance
	m, err := migration.NewMigration(*sqliteDBPath, postgresConnStr)
	if err != nil {
		log.Fatalf("Error creating migration: %v", err)
	}
	defer m.Close()

	// Create PostgreSQL schema
	log.Println("Creating PostgreSQL schema...")
	if err := m.CreatePostgresSchema(); err != nil {
		log.Fatalf("Error creating schema: %v", err)
	}

	// Migrate data
	log.Println("Starting data migration...")
	if err := m.MigrateData(); err != nil {
		log.Fatalf("Error migrating data: %v", err)
	}

	log.Println("Migration completed successfully!")
}
