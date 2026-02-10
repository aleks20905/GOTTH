package handlers

import (
	"goth/internal/templates"
	"net/http"
)

type AccountHandler struct{}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{}
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.AccountPage()
	err := templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
