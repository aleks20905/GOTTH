package templates

import "goth/internal/store"

// ComputerRow is one agent's status as shown on the computers page.
// Offline rows carry the last known version and OS, "—" for uptime,
// and a humanized last-seen timestamp.
type ComputerRow struct {
	ID       uint
	URL      string
	Name     string
	Version  string
	OS       string
	Uptime   string
	LastSeen string
	Online   bool
	Updating bool
	Verified bool
}

// TaskRow is one update task with its resolved human-readable target.
type TaskRow struct {
	Task   store.UpdateTask
	Target string
}

// HomeStats aggregates fleet numbers for the dashboard.
type HomeStats struct {
	Total        int
	Online       int
	Updating     int
	PendingTasks int
}
