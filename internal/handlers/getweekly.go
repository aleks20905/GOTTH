package handlers

import (
	"goth/internal/templates"
	"net/http"
)

type weeklyHandLer struct{}

func NewWeeklyHandler() *weeklyHandLer {
	return &weeklyHandLer{}
}

func (h *weeklyHandLer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.Weekly()
	err := templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
