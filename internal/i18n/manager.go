package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manager handles translations for different languages
type Manager struct {
	defaultLang  string
	translations map[string]map[string]string
}

// NewManager creates a new i18n manager with the specified default language
func NewManager(defaultLang string) *Manager {
	return &Manager{
		defaultLang:  defaultLang,
		translations: make(map[string]map[string]string),
	}
}

// LoadTranslations loads all translation files from the specified directory
func (m *Manager) LoadTranslations(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read translations directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			lang := strings.TrimSuffix(entry.Name(), ".json")

			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read translation file %s: %w", entry.Name(), err)
			}

			var translations map[string]string
			if err := json.Unmarshal(data, &translations); err != nil {
				return fmt.Errorf("failed to parse translation file %s: %w", entry.Name(), err)
			}

			m.translations[lang] = translations
		}
	}

	return nil
}

// Translate returns the translation for the given key in the specified language
func (m *Manager) Translate(lang, key string) string {
	// Try requested language
	if translations, ok := m.translations[lang]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}

	// Fallback to default language
	if translations, ok := m.translations[m.defaultLang]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}

	// Return key if no translation found
	return key
}

// GetDefaultLang returns the default language
func (m *Manager) GetDefaultLang() string {
	return m.defaultLang
}

// GetAvailableLanguages returns a list of available languages
func (m *Manager) GetAvailableLanguages() []string {
	langs := make([]string, 0, len(m.translations))
	for lang := range m.translations {
		langs = append(langs, lang)
	}
	return langs
}
