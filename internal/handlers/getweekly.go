package handlers

import (
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
)

type weeklyHandLer struct {
	scheduleStore store.ScheduleStore
}

type getWeeklyHandlerParams struct {
	ScheduleStore store.ScheduleStore
}

func NewWeeklyHandler(params getWeeklyHandlerParams) *weeklyHandLer {
	return &weeklyHandLer{
		scheduleStore: params.ScheduleStore,
	}
}

func (h *weeklyHandLer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.Weekly()
	err := templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
