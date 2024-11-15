// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778

package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"
	"time"
)

func Weekly() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<p>weekly things test </p><div class=\"calendar\"><div class=\"timeline\"><div class=\"spacer\"></div><div class=\"time-marker\">8:00</div><div class=\"time-marker\">9:00</div><div class=\"time-marker\">10:00</div><div class=\"time-marker\">11:00</div><div class=\"time-marker\">12:00</div><div class=\"time-marker\">13:00</div><div class=\"time-marker\">14:00</div><div class=\"time-marker\">15:00</div><div class=\"time-marker\">16:00</div><div class=\"time-marker\">17:00</div><div class=\"time-marker\">18:00</div></div><div class=\"days\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for day, dayEvents := range replace_after() {
			var templ_7745c5c3_Var2 = []any{"day", day}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var2...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var2).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/weekly.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div class=\"date\"><p class=\"date-day\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(day)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/weekly.templ`, Line: 31, Col: 45}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p></div><div class=\"events\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, event := range dayEvents {
				var templ_7745c5c3_Var5 = []any{"event", event.EndstartEvent, "securities"}
				templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var5...)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var6 string
				templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var5).String())
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/weekly.templ`, Line: 1, Col: 0}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><p class=\"title\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var7 string
				templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(event.Title)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/weekly.templ`, Line: 36, Col: 57}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p><p class=\"time\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var8 string
				templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(event.TimeRange)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/templates/weekly.templ`, Line: 37, Col: 62}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

// Event struct to hold event data
type Event struct {
	Title         string
	Start         time.Time
	End           time.Time
	TimeRange     string
	DayOfWeek     string // Day of the week (e.g., Monday, Tuesday)
	EndstartEvent string
}

// Function to format time in "HH:MM - HH:MM" 24-hour format
func formatTimeRange(start, end time.Time) string {
	startTime := start.Format("15:04")
	endTime := end.Format("15:04")
	return fmt.Sprintf("%s-%s", startTime, endTime) // start-2 end-5
}

func formatEndStart(start, end time.Time) string {
	startTime := start.Format("15")
	endTime := end.Format("15")
	return fmt.Sprintf("start-%s end-%s", startTime, endTime)
}

// Helper function to parse time from string
func parseTime(timeStr string) time.Time {
	layout := "2006-01-02 15:04:05"
	parsedTime, _ := time.Parse(layout, timeStr)
	return parsedTime
}

// Function to group events by their day of the week
func groupEventsByDay(events []Event) map[string][]Event {
	eventsByDay := make(map[string][]Event)

	// Loop through events to populate TimeRange and DayOfWeek fields
	for _, event := range events {

		event.EndstartEvent = formatEndStart(event.Start, event.End)
		event.TimeRange = formatTimeRange(event.Start, event.End)
		event.DayOfWeek = event.Start.Format("Monday") // Get the full name of the day
		eventsByDay[event.DayOfWeek] = append(eventsByDay[event.DayOfWeek], event)
	}

	return eventsByDay
}

// Function to generate a structured list of events grouped by day
func replace_after() map[string][]Event {
	events := []Event{
		{Title: "Event 1", Start: parseTime("2024-10-21 08:00:00"), End: parseTime("2024-10-21 11:00:00")},
		{Title: "Event 2", Start: parseTime("2024-10-21 11:00:00"), End: parseTime("2024-10-21 13:00:00")},
		{Title: "Event 3", Start: parseTime("2024-10-22 10:00:00"), End: parseTime("2024-10-22 12:00:00")},
		{Title: "Event 4", Start: parseTime("2024-10-23 14:00:00"), End: parseTime("2024-10-23 16:00:00")},
		{Title: "Event 5", Start: parseTime("2024-10-21 14:00:00"), End: parseTime("2024-10-21 16:00:00")},
		{Title: "Event 6", Start: parseTime("2024-10-22 12:00:00"), End: parseTime("2024-10-22 15:00:00")},
		{Title: "Event 7", Start: parseTime("2024-10-24 08:00:00"), End: parseTime("2024-10-24 11:00:00")},
		{Title: "Event 8", Start: parseTime("2024-10-25 11:00:00"), End: parseTime("2024-10-25 13:00:00")},
		{Title: "Event 9", Start: parseTime("2024-10-25 14:00:00"), End: parseTime("2024-10-25 16:00:00")},
	}

	// Group events by their day of the week
	return groupEventsByDay(events)
}

var _ = templruntime.GeneratedTemplate
