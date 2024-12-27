package handlers

import (
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
)

type weeklyListHandLer struct {
	scheduleStore store.ScheduleStore
}

type GetWeeklyListHandlerParams struct {
	ScheduleStore store.ScheduleStore
}

func NewWeeklyListHandler(params GetWeeklyListHandlerParams) *weeklyListHandLer {
	return &weeklyListHandLer{
		scheduleStore: params.ScheduleStore,
	}
}

func (h *weeklyListHandLer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// course, err := convertCourse(r.URL.Query().Get("course"))
	// if err != nil {
	// 	http.Error(w, "Error course parse problem", http.StatusInternalServerError)
	// 	return
	// }
	// spec := r.URL.Query().Get("spec")
	// groupName := r.URL.Query().Get("group_name")

	idk, err := h.scheduleStore.GetAllscheduleUrls()
	if err != nil {
		http.Error(w, "Error loading schedule"+err.Error(), http.StatusInternalServerError)
		return
	}

	c := templates.WeeklyList(*idk)
	err = templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
