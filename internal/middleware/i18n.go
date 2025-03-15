package middleware

import (
	"net/http"

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

			// Set default language if none specified
			if lang == "" {
				lang = "en"
			}

			// Set language cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "lang",
				Value:    lang,
				Path:     "/",
				MaxAge:   365 * 24 * 60 * 60, // 1 year
				HttpOnly: true,
				Secure:   r.TLS != nil,
			})

			// Update request context with language
			ctx := i18n.WithLang(r.Context(), lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
