package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"time"

	"expensemanager/internal/database"
	"expensemanager/internal/handlers"
	"expensemanager/internal/middleware"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	// Initialize database
	db, err := database.NewDB("./db/expenses.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Initialize(); err != nil {
		log.Fatal(err)
	}

	// Template functions
	funcMap := template.FuncMap{
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
		"now": time.Now,
	}

	// Parse templates with functions
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html"))

	// Initialize handlers
	h := handlers.NewHandler(db, tmpl)

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

	// Wrap the mux with security headers middleware
	handler := middleware.SecurityHeaders(mux)

	log.Println("Server starting at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
