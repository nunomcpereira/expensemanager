package middleware

import (
	"expensemanager/internal/i18n"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// Middleware represents a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware in order
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// Logger middleware logs request details
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Log request details
		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.RequestURI,
			rw.status,
			time.Since(start),
		)
	})
}

// I18n middleware sets the language in the context
func I18n(manager *i18n.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get language from query parameter or session
			lang := r.URL.Query().Get("lang")
			if lang == "" {
				// Try to get language from session
				session, err := r.Cookie("session")
				if err == nil {
					lang = session.Value
				}
			}

			// Set default language if none specified
			if lang == "" {
				lang = "en"
			}

			// Store language in context
			ctx := i18n.WithLang(r.Context(), lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Recovery middleware recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				log.Printf("panic: %v\n%s", err, debug.Stack())

				// Return 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
