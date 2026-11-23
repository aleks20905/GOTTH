package handlers

import (
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UsersHandlerParams struct {
	UserStore store.UserStore
	LogStore  store.LogStore
}

type UsersHandler struct {
	userStore store.UserStore
	logStore  store.LogStore
}

func NewUsersHandler(params UsersHandlerParams) *UsersHandler {
	return &UsersHandler{
		userStore: params.UserStore,
		logStore:  params.LogStore,
	}
}

func (h *UsersHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return "", false
	}
	if !middleware.IsAdmin(r.Context()) {
		http.Redirect(w, r, "/", http.StatusFound)
		return "", false
	}
	return user.Email, true
}

func (h *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	users, err := h.userStore.ListUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := templates.Users(users, email)
	if err := templates.Layout(c, "Users").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *UsersHandler) GrantAdmin(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.userStore.SetAdmin(uint(id), true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "users", "Admin granted.", strconv.FormatUint(id, 10), email)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *UsersHandler) RevokeAdmin(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.userStore.SetAdmin(uint(id), false); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "users", "Admin revoked.", strconv.FormatUint(id, 10), email)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
