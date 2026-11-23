package handlers

import (
	"goth/internal/middleware"
	"goth/internal/templates"
	"net/http"
)

type AccountHandler struct{}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{}
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	c := templates.Account(user.Email, r.URL.Query().Get("msg"))
	err := templates.Layout(c, "Account").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
