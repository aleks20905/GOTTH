package handlers

import (
	"goth/internal/hash"
	"goth/internal/middleware"
	"goth/internal/store"
	"net/http"
	"strings"
)

type PostAccountHandlerParams struct {
	UserStore    store.UserStore
	PasswordHash hash.PasswordHash
	LogStore     store.LogStore
}

type PostAccountHandler struct {
	userStore    store.UserStore
	passwordHash hash.PasswordHash
	logStore     store.LogStore
}

func NewPostAccountHandler(params PostAccountHandlerParams) *PostAccountHandler {
	return &PostAccountHandler{
		userStore:    params.UserStore,
		passwordHash: params.PasswordHash,
		logStore:     params.LogStore,
	}
}

func (h *PostAccountHandler) currentUser(w http.ResponseWriter, r *http.Request) (*store.User, bool) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, false
	}
	return user, true
}

func (h *PostAccountHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" || !strings.Contains(email, "@") {
		http.Redirect(w, r, "/account?msg=email-invalid", http.StatusSeeOther)
		return
	}

	if err := h.userStore.UpdateEmail(user.ID, email); err != nil {
		http.Redirect(w, r, "/account?msg=email-taken", http.StatusSeeOther)
		return
	}

	writeLog(h.logStore, "info", "account", "Email changed.", email, user.Email)
	http.Redirect(w, r, "/account?msg=email-updated", http.StatusSeeOther)
}

func (h *PostAccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}

	current := r.FormValue("current_password")
	next := r.FormValue("new_password")

	if len(next) < 8 {
		http.Redirect(w, r, "/account?msg=password-short", http.StatusSeeOther)
		return
	}

	valid, err := h.passwordHash.ComparePasswordAndHash(current, user.Password)
	if err != nil || !valid {
		http.Redirect(w, r, "/account?msg=password-wrong", http.StatusSeeOther)
		return
	}

	hashed, err := h.passwordHash.GenerateFromPassword(next)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.userStore.UpdatePassword(user.ID, hashed); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "account", "Password changed.", "", user.Email)
	http.Redirect(w, r, "/account?msg=password-updated", http.StatusSeeOther)
}
