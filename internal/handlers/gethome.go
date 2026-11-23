package handlers

import (
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
)

type HomeDataParams struct {
	DeviceStore  store.DeviceStore
	TaskStore    store.TaskStore
	LogStore     store.LogStore
	StatusGetter StatusGetter
}

type HomeHandler struct {
	deviceStore  store.DeviceStore
	taskStore    store.TaskStore
	logStore     store.LogStore
	statusGetter StatusGetter
}

func NewHomeHandler(params HomeDataParams) *HomeHandler {
	return &HomeHandler{
		deviceStore:  params.DeviceStore,
		taskStore:    params.TaskStore,
		logStore:     params.LogStore,
		statusGetter: params.StatusGetter,
	}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserKey).(*store.User)

	if !ok {
		c := templates.GuestIndex()

		err := templates.Layout(c, "RemoteDesk").Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}

		return
	}

	devices, err := h.deviceStore.ListDevices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows := buildComputerRows(h.statusGetter, devices)

	pending, err := h.taskStore.ListTasksByStatus("pending")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	running, err := h.taskStore.ListTasksByStatus("running")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	entries, err := h.logStore.ListEntries(5)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stats := templates.HomeStats{Total: len(rows)}
	for _, row := range rows {
		if row.Updating {
			stats.Updating++
		} else if row.Online {
			stats.Online++
		}
	}
	stats.PendingTasks = len(pending) + len(running)

	c := templates.Index(user.Email, stats, entries)
	err = templates.Layout(c, "RemoteDesk").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
