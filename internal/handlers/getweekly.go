package handlers

import (
	"fmt"
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

	spec := r.URL.Query().Get("spec")
	groupName := r.URL.Query().Get("group_name")
	course, err := convertCourse(r.URL.Query().Get("course")) // converts the string to uint
	if err != nil {
		http.Error(w, "Error course parse problem", http.StatusInternalServerError)
		return
	}

	// courses, err := h.scheduleStore.GetCourses()
	// if err != nil {
	// 	http.Error(w, "Error loading course"+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// specs, err := h.scheduleStore.GetSpecs()
	// if err != nil {
	// 	http.Error(w, "Error loading Specs"+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// groupNames, err := h.scheduleStore.GetGroupNames()
	// if err != nil {
	// 	http.Error(w, "Error loading GroupNames"+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// gets the schedule
	schedule, err := h.scheduleStore.GetSchedules(course, spec, groupName)
	if err != nil {
		http.Error(w, "Error loading schedule"+err.Error(), http.StatusInternalServerError)
		return
	}

	// converts the shedule to daySchedule
	daySchedules, err := scheduleToDays(*schedule)
	if err != nil {
		http.Error(w, "Error conventing to dayschedule"+err.Error(), http.StatusInternalServerError)
		return
	}

	// render the template
	c := templates.Weekly(daySchedules)
	// c := templates.Weekly(*courses, *specs, *groupNames, daySchedules)
	err = templates.Layout(c, "My website").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// scheduleToDays takes []Schedule or nil and return []DaySchedules and it aways contain monday to fri
func scheduleToDays(schedules []store.Schedule) ([]store.DayScheduels, error) {
	weekdaysInOrder := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

	// Return an empty week schedule if no schedules are provided
	if len(schedules) == 0 {
		weekSchedule := make([]store.DayScheduels, len(weekdaysInOrder))
		for i, day := range weekdaysInOrder {
			weekSchedule[i] = store.DayScheduels{
				Day:     day,
				Shedule: []store.Schedule{},
			}
		}
		return weekSchedule, nil
	}

	weekSchedule := make([]store.DayScheduels, len(weekdaysInOrder))
	for i, day := range weekdaysInOrder {
		weekSchedule[i] = store.DayScheduels{
			Day:     day,
			Shedule: []store.Schedule{},
		}
	}

	for _, schedule := range schedules {
		// Get the weekday string (Monday, Tuesday, etc.)
		weekday := schedule.Start.Weekday().String()

		// Check if the weekday is valid
		weekdayFound := false
		for i := range weekSchedule {
			if weekSchedule[i].Day == weekday {
				weekSchedule[i].Shedule = append(weekSchedule[i].Shedule, schedule)
				weekdayFound = true
				break
			}
		}

		// Return an error if the weekday from the schedule is not in the predefined weekday list
		if !weekdayFound {
			return nil, fmt.Errorf("invalid weekday found in schedule: %s", weekday)
		}
	}

	// Successfully return the weekSchedule
	return weekSchedule, nil
}

func convertCourse(courseStr string) (uint, error) {
	if courseStr == "" {
		return 0, nil
	}

	course, err := strconv.ParseUint(courseStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(course), nil
}
