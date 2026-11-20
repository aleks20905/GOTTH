package handlers

import (
	"fmt"
	"goth/internal/agent"
	"goth/internal/middleware"
	"goth/internal/templates"
	"net/http"
)

type StatusGetter interface {
	GetStatus(baseURL string) (*agent.Status, error)
}

type ComputersHandlerParams struct {
	Agents       []string
	StatusGetter StatusGetter
}

type ComputersHandler struct {
	agents       []string
	statusGetter StatusGetter
}

func NewComputersHandler(params ComputersHandlerParams) *ComputersHandler {
	return &ComputersHandler{
		agents:       params.Agents,
		statusGetter: params.StatusGetter,
	}
}

func (h *ComputersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	rows := make([]templates.ComputerRow, 0, len(h.agents))
	for _, url := range h.agents {
		row := templates.ComputerRow{URL: url, Name: "—", Version: "—", OS: "—", Uptime: "—"}
		if status, err := h.statusGetter.GetStatus(url); err == nil {
			row.Name = status.Name
			row.Version = status.Version
			row.OS = status.OS
			row.Uptime = humanizeUptime(status.UptimeSeconds)
			row.Online = true
		}
		rows = append(rows, row)
	}

	c := templates.Computers(rows)
	err := templates.Layout(c, "Компютри").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func humanizeUptime(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	switch {
	case days > 0:
		return fmt.Sprintf("%d д %d ч", days, hours)
	case hours > 0:
		return fmt.Sprintf("%d ч %d мин", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%d мин", minutes)
	default:
		return fmt.Sprintf("%d сек", seconds)
	}
}
