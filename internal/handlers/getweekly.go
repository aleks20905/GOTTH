package handlers

import (
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
)

type weeklyHandLer struct {
	scheduleStore store.ScheduleStore
}

type GetWeeklyHandlerParams struct {
	ScheduleStore store.ScheduleStore
}

func NewWeeklyHandler(params GetWeeklyHandlerParams) *weeklyHandLer {
	return &weeklyHandLer{
		scheduleStore: params.ScheduleStore,
	}
}

func (h *weeklyHandLer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	user, err := h.scheduleStore.GetSchedule(1, "idk", "idk")
	if err != nil {
		http.Error(w, "Error loading schedule", http.StatusInternalServerError)
		return
	}
	if user != nil {
		id = "it works"
	}
	c := templates.Weekly(id)
	err = templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
