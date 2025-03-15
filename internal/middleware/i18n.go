package middleware

import (
	"net/http"
	"time"

	"expensemanager/internal/i18n"
)

// WithI18n creates a new i18n middleware with the given manager
func WithI18n(manager *i18n.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get language from query parameter or cookie
			lang := r.URL.Query().Get("lang")
			if lang == "" {
				if cookie, err := r.Cookie("lang"); err == nil {
					lang = cookie.Value
				}
			}

			// Validate language or set default
			availableLangs := manager.GetAvailableLanguages()
			if lang == "" || !containsLang(availableLangs, lang) {
				lang = manager.GetDefaultLang()
			}

			// Set language cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "lang",
				Value:    lang,
				Path:     "/",
				Expires:  time.Now().AddDate(1, 0, 0), // 1 year
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   r.TLS != nil,
			})

			// Update request context with language
			ctx := i18n.WithLang(r.Context(), lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// containsLang checks if a language is in the available languages slice
func containsLang(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
