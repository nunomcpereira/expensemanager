package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type sessionStoreKey struct{}

// WithSessionStore creates middleware that adds the session store to the request context
func WithSessionStore(store sessions.Store) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), sessionStoreKey{}, store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSessionStore retrieves the session store from the context
func GetSessionStore(ctx context.Context) (sessions.Store, bool) {
	store, ok := ctx.Value(sessionStoreKey{}).(sessions.Store)
	return store, ok
}
