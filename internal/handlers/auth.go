package handlers

import (
	"context"
	"expensemanager/internal/database"
	"expensemanager/internal/i18n"
	"expensemanager/internal/models"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// AuthTemplateData holds data to be passed to auth templates
type AuthTemplateData struct {
	Lang               string
	AvailableLanguages []string
	Error              string
	User               *models.User
}

type AuthHandler struct {
	db    *database.DB
	tmpl  *template.Template
	store sessions.Store
	i18n  *i18n.Manager
}

func NewAuthHandler(db *database.DB, tmpl *template.Template, store sessions.Store) *AuthHandler {
	return &AuthHandler{
		db:    db,
		tmpl:  tmpl,
		store: store,
	}
}

// UpdateI18n updates the i18n manager
func (h *AuthHandler) UpdateI18n(manager *i18n.Manager) {
	h.i18n = manager
}

// GetTemplateData prepares common template data
func (h *AuthHandler) GetTemplateData(r *http.Request) *AuthTemplateData {
	data := &AuthTemplateData{
		Lang:               "en",                           // Default language
		AvailableLanguages: h.i18n.GetAvailableLanguages(), // Get available languages from i18n manager
	}

	// Get language from session if available
	session, _ := h.store.Get(r, "session")
	if lang, ok := session.Values["lang"].(string); ok {
		data.Lang = lang
	}

	return data
}

type userIDKey struct{}

func SetUserIDContext(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey{}).(int64)
	return userID, ok
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData(r)

	if r.Method == http.MethodPost {
		form := &models.LoginForm{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
			Language: r.FormValue("language"),
		}

		user, err := h.db.AuthenticateUser(form.Email, form.Password)
		if err != nil {
			data.Error = "Invalid email or password"
			if err := h.tmpl.ExecuteTemplate(w, "login", data); err != nil {
				log.Printf("Error executing login template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Set session
		session, _ := h.store.Get(r, "session")
		session.Values["user_id"] = user.ID
		session.Values["user_email"] = user.Email
		session.Values["user_name"] = user.Name
		session.Values["language"] = form.Language
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "login", data); err != nil {
		log.Printf("Error executing login template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	data := h.GetTemplateData(r)

	if r.Method == http.MethodPost {
		form := &models.RegisterForm{
			Email:           r.FormValue("email"),
			Password:        r.FormValue("password"),
			ConfirmPassword: r.FormValue("confirm_password"),
			Name:            r.FormValue("name"),
		}

		// Validate form
		if form.Password != form.ConfirmPassword {
			data.Error = "Passwords do not match"
			h.tmpl.ExecuteTemplate(w, "register", data)
			return
		}

		// Check if user already exists
		existingUser, err := h.db.GetUserByEmail(form.Email)
		if err != nil {
			data.Error = "An error occurred"
			h.tmpl.ExecuteTemplate(w, "register", data)
			return
		}
		if existingUser != nil {
			data.Error = "Email already registered"
			h.tmpl.ExecuteTemplate(w, "register", data)
			return
		}

		// Create user
		user := &models.User{
			Email: form.Email,
			Name:  form.Name,
		}
		err = h.db.CreateUser(user, form.Password)
		if err != nil {
			data.Error = "Failed to create user"
			h.tmpl.ExecuteTemplate(w, "register", data)
			return
		}

		// Set session
		session, _ := h.store.Get(r, "session")
		session.Values["user_id"] = user.ID
		session.Values["user_email"] = user.Email
		session.Values["user_name"] = user.Name
		if err := session.Save(r, w); err != nil {
			log.Printf("Error saving session: %v", err)
			data.Error = "Failed to create session"
			h.tmpl.ExecuteTemplate(w, "register", data)
			return
		}

		// Set user ID in context and redirect
		r = r.WithContext(SetUserIDContext(r.Context(), user.ID))
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	h.tmpl.ExecuteTemplate(w, "register", data)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.store.Get(r, "session")
		userID, ok := session.Values["user_id"].(int64)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		r = r.WithContext(SetUserIDContext(r.Context(), userID))
		next(w, r)
	}
}

// HandleLanguage handles language changes for both authenticated and unauthenticated users
func (h *AuthHandler) HandleLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	if lang == "" {
		http.Error(w, "Language not specified", http.StatusBadRequest)
		return
	}

	// Store language preference in session
	session, _ := h.store.Get(r, "session")
	session.Values["lang"] = lang
	session.Save(r, w)

	// Redirect back to the previous page
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}
