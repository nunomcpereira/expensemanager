package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"expensemanager/internal/database"
	"expensemanager/internal/handlers"
	"expensemanager/internal/i18n"
	"expensemanager/internal/middleware"

	_ "github.com/lib/pq" // PostgreSQL driver
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	// Build PostgreSQL connection string from environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "expensemanager"
	}
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	// Format PostgreSQL connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)

	// Initialize database
	db, err := database.NewDB(connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Initialize(); err != nil {
		log.Fatal(err)
	}

	// Initialize i18n manager
	i18nManager := i18n.NewManager("en")
	if err := i18nManager.LoadTranslations("internal/i18n/locales"); err != nil {
		log.Fatal(err)
	}

	// Template functions
	funcMap := template.FuncMap{
		// Math operations
		"add": func(a, b float64) float64 {
			return a + b
		},
		"sub": func(a, b float64) float64 {
			return a - b
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		// Comparison functions
		"gt": func(a, b float64) bool {
			return a > b
		},
		"lt": func(a, b float64) bool {
			return a < b
		},
		"eq": func(a, b float64) bool {
			return a == b
		},
		"eqs": func(a, b string) bool {
			return a == b
		},
		// Time functions
		"now": time.Now,
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		// Money formatting
		"formatMoney": func(amount float64) string {
			return fmt.Sprintf("$%.2f", amount)
		},
		// Translation function
		"t": func(lang, key string) string {
			return i18nManager.Translate(lang, key)
		},
		// String manipulation
		"lower": strings.ToLower,
	}

	// Parse templates with functions
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html"))

	// Initialize handlers
	h := handlers.NewHandler(db, tmpl)
	h.UpdateI18n(i18nManager)

	// Create a new mux for routing
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Routes
	mux.HandleFunc("/", h.HandleIndex)
	mux.HandleFunc("/expenses", h.HandleExpenses)
	mux.HandleFunc("/expenses/add", h.HandleAddExpense)
	mux.HandleFunc("/expenses/delete", h.HandleDeleteExpense)
	mux.HandleFunc("/summary", h.HandleSummary)
	mux.HandleFunc("/reports", h.HandleReports)
	mux.HandleFunc("/api/monthly-totals", h.HandleMonthlyTotals)
	mux.HandleFunc("/api/category-totals", h.HandleCategoryTotals)
	mux.HandleFunc("/admin", h.HandleAdmin)
	mux.HandleFunc("/admin/clear-expenses", h.HandleClearExpenses)
	mux.HandleFunc("/admin/download-expenses", h.HandleDownloadExpenses)
	mux.HandleFunc("/admin/upload-expenses", h.HandleUploadExpenses)

	// Wrap the mux with middleware
	handler := middleware.Chain(
		mux,
		middleware.SecurityHeaders,
		middleware.WithI18n(i18nManager),
	)

	log.Println("Server starting at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
