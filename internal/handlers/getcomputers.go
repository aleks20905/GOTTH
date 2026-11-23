package handlers

import (
	"fmt"
	"goth/internal/agent"
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"strconv"
	"time"
)

type StatusGetter interface {
	GetStatus(baseURL string) (*agent.Status, error)
}

type ComputersHandlerParams struct {
	DeviceStore  store.DeviceStore
	GroupStore   store.GroupStore
	StatusGetter StatusGetter
}

type ComputersHandler struct {
	deviceStore  store.DeviceStore
	groupStore   store.GroupStore
	statusGetter StatusGetter
}

func NewComputersHandler(params ComputersHandlerParams) *ComputersHandler {
	return &ComputersHandler{
		deviceStore:  params.DeviceStore,
		groupStore:   params.GroupStore,
		statusGetter: params.StatusGetter,
	}
}

func (h *ComputersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	devices, err := h.deviceStore.ListDevices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	groups, err := h.groupStore.ListGroups()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var groupFilter *uint
	if raw := r.URL.Query().Get("group"); raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 32); err == nil {
			gid := uint(id)
			groupFilter = &gid
		}
	}

	filtered := devices[:0]
	for _, device := range devices {
		if groupFilter == nil || (device.GroupID != nil && *device.GroupID == *groupFilter) {
			filtered = append(filtered, device)
		}
	}

	c := templates.Computers(buildComputerRows(h.statusGetter, filtered), groups, groupFilter)
	err = templates.Layout(c, "Computers").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func buildComputerRows(getter StatusGetter, devices []store.Device) []templates.ComputerRow {
	rows := make([]templates.ComputerRow, 0, len(devices))
	for _, device := range devices {
		row := templates.ComputerRow{ID: device.ID, URL: device.URL, Name: device.Name, Version: orDash(device.Version), OS: orDash(device.OS), Uptime: "—", LastSeen: humanizeLastSeen(device.LastSeen), Verified: device.Verified}
		if status, err := getter.GetStatus(device.URL); err == nil {
			if status.Name != "" {
				row.Name = status.Name
			}
			row.Version = status.Version
			row.OS = status.OS
			row.Uptime = humanizeUptime(status.UptimeSeconds)
			row.Online = true
			row.Updating = status.Updating
		}
		rows = append(rows, row)
	}
	return rows
}

func orDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func humanizeLastSeen(lastSeen time.Time) string {
	if lastSeen.IsZero() {
		return "—"
	}
	ago := time.Since(lastSeen)
	switch {
	case ago < time.Minute:
		return "just now"
	case ago < time.Hour:
		return fmt.Sprintf("%d min ago", int(ago.Minutes()))
	case ago < 24*time.Hour:
		return fmt.Sprintf("%d h ago", int(ago.Hours()))
	case ago < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(ago.Hours()/24))
	default:
		return lastSeen.Format("02.01.2006")
	}
}

func humanizeUptime(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	switch {
	case days > 0:
		return fmt.Sprintf("%d d %d h", days, hours)
	case hours > 0:
		return fmt.Sprintf("%d h %d min", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%d min", minutes)
	default:
		return fmt.Sprintf("%d sec", seconds)
	}
}
