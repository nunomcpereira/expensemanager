package i18n

import (
	"context"
)

type contextKey string

const langKey contextKey = "lang"

// WithLang returns a new context with the language value
func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, langKey, lang)
}

// GetLang retrieves the language from the context
func GetLang(ctx context.Context) string {
	if lang, ok := ctx.Value(langKey).(string); ok {
		return lang
	}
	return "en" // Default to English if no language is set
}
