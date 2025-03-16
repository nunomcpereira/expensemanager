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

	"github.com/gorilla/sessions"
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
	log.Printf("Loading i18n translations...")
	i18nManager := i18n.NewManager("en")
	if err := i18nManager.LoadTranslations("internal/i18n/locales"); err != nil {
		log.Fatalf("Failed to load translations: %v", err)
	}
	log.Printf("Available languages: %v", i18nManager.GetAvailableLanguages())

	// Initialize session store
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		sessionKey = "your-secret-key-change-me" // Change this in production
	}
	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
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
	log.Printf("Loading templates from embedded filesystem...")
	templates, err := templatesFS.ReadDir("templates")
	if err != nil {
		log.Fatalf("Failed to read templates directory: %v", err)
	}
	for _, template := range templates {
		log.Printf("Found template: %s", template.Name())
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	log.Printf("Templates loaded successfully")

	// Initialize handlers
	h := handlers.NewHandler(db, tmpl, store)
	h.UpdateI18n(i18nManager)

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler(db, tmpl, store)
	authHandler.UpdateI18n(i18nManager)

	// Create a new mux for routing
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Auth routes
	mux.HandleFunc("/login", authHandler.HandleLogin)
	mux.HandleFunc("/register", authHandler.HandleRegister)
	mux.HandleFunc("/logout", authHandler.HandleLogout)

	// Protected routes
	mux.HandleFunc("/", authHandler.RequireAuth(h.HandleIndex))
	mux.HandleFunc("/expenses", authHandler.RequireAuth(h.HandleExpenses))
	mux.HandleFunc("/expenses/add", authHandler.RequireAuth(h.HandleAddExpense))
	mux.HandleFunc("/expenses/delete", authHandler.RequireAuth(h.HandleDeleteExpense))
	mux.HandleFunc("/summary", authHandler.RequireAuth(h.HandleSummary))
	mux.HandleFunc("/reports", authHandler.RequireAuth(h.HandleReports))
	mux.HandleFunc("/api/monthly-totals", authHandler.RequireAuth(h.HandleMonthlyTotals))
	mux.HandleFunc("/api/category-totals", authHandler.RequireAuth(h.HandleCategoryTotals))
	mux.HandleFunc("/admin", authHandler.RequireAuth(h.HandleAdmin))
	mux.HandleFunc("/admin/clear-expenses", authHandler.RequireAuth(h.HandleClearExpenses))
	mux.HandleFunc("/admin/download-expenses", authHandler.RequireAuth(h.HandleDownloadExpenses))
	mux.HandleFunc("/admin/upload-expenses", authHandler.RequireAuth(h.HandleUploadExpenses))

	// Language route
	mux.HandleFunc("/language", authHandler.HandleLanguage)

	// Wrap the mux with middleware
	handler := middleware.Chain(mux,
		middleware.Logger,
		middleware.WithSessionStore(store),
		middleware.I18n(i18nManager),
		middleware.Recovery,
	)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
