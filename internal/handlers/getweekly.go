package handlers

import (
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"strconv"
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
	courseStr := r.URL.Query().Get("course")
	if courseStr == "" {
		http.Error(w, "Missing 'course' query parameter"+courseStr, http.StatusBadRequest)
		return
	}

	course, err := strconv.ParseUint(courseStr, 10, 64)
	if err != nil {
		http.Error(w, "Error url query problem"+courseStr, http.StatusInternalServerError)
		return
	}
	uintCourse := uint(course)

	spec := r.URL.Query().Get("spec")
	groupName := r.URL.Query().Get("group_name")

	schedule, err := h.scheduleStore.GetSchedule(uintCourse, spec, groupName)
	if err != nil {
		http.Error(w, "Error loading schedule", http.StatusInternalServerError)
		return
	}

	daySchedules := scheduleToDays(*schedule)

	c := templates.Weekly(daySchedules)
	err = templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func scheduleToDays(schedules []store.Schedule) []store.DayScheduels {
	weekdaysInOrder := []string{
		"Monday", "Tuesday", "Wednesday", "Thursday", "Friday",
	}

	weekSchedule := make([]store.DayScheduels, len(weekdaysInOrder))
	for i, day := range weekdaysInOrder {
		weekSchedule[i] = store.DayScheduels{
			Day:     day,
			Shedule: []store.Schedule{},
		}
	}

	for _, schedule := range schedules {
		weekday := schedule.Start.Weekday().String() // Get the weekday string (Monday, Tuesday, etc.)

		for i := range weekSchedule {
			if weekSchedule[i].Day == weekday { // Use the Day field from store.DayScheduels
				weekSchedule[i].Shedule = append(weekSchedule[i].Shedule, schedule)
				break
			}
		}
	}

	// for _, ds := range weekSchedule {
	// 	fmt.Printf("Day: %s, Schedules: %+v\n", ds.Day, ds.Shedule)
	// }

	return weekSchedule
}
