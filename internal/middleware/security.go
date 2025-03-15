package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CSP header
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com; "+
				"style-src-elem 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com; "+
				"img-src 'self' data: blob:; "+
				"worker-src 'self' blob:; "+
				"font-src 'self' https://cdn.jsdelivr.net https://unpkg.com; "+
				"connect-src 'self';")

		// Add other security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}
