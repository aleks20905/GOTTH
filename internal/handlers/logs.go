package handlers

import (
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
)

type LogsHandlerParams struct {
	LogStore store.LogStore
}

type LogsHandler struct {
	logStore store.LogStore
}

func NewLogsHandler(params LogsHandlerParams) *LogsHandler {
	return &LogsHandler{
		logStore: params.LogStore,
	}
}

func (h *LogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	entries, err := h.logStore.ListEntries(200)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := templates.Logs(entries)
	if err := templates.Layout(c, "Logs").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
