package handlers

import (
	"encoding/json"
	"goth/internal/agent"
	"goth/internal/store"
	"net/http"
	"time"
)

// RegisterHandler pairs agents that phone home. Machine-to-machine endpoint,
// the shared AGENT_TOKEN is the auth, so it lives outside the login group.
type RegisterHandlerParams struct {
	DeviceStore store.DeviceStore
	TaskStore   store.TaskStore
	LogStore    store.LogStore
	AgentToken  string
}

type RegisterHandler struct {
	deviceStore store.DeviceStore
	taskStore   store.TaskStore
	logStore    store.LogStore
	agentToken  string
}

func NewRegisterHandler(params RegisterHandlerParams) *RegisterHandler {
	return &RegisterHandler{
		deviceStore: params.DeviceStore,
		taskStore:   params.TaskStore,
		logStore:    params.LogStore,
		agentToken:  params.AgentToken,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var reg agent.Registration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.agentToken == "" || reg.URL == "" || !agent.TokensMatch(h.agentToken, reg.Token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	device, err := h.deviceStore.UpsertDevice(&store.Device{
		Name:     reg.Name,
		URL:      reg.URL,
		Version:  reg.Version,
		OS:       reg.OS,
		Verified: true,
		LastSeen: time.Now(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.completeMatchingTasks(device, reg.Version)

	w.WriteHeader(http.StatusNoContent)
}

// completeMatchingTasks closes running update tasks whose target version the
// agent now reports. The heartbeat is the completion signal.
func (h *RegisterHandler) completeMatchingTasks(device *store.Device, version string) {
	if version == "" {
		return
	}

	running, err := h.taskStore.ListTasksByStatus("running")
	if err != nil {
		return
	}

	for _, task := range running {
		if task.PackageVersion != version {
			continue
		}
		if task.DeviceID != nil && *task.DeviceID != device.ID {
			continue
		}
		if task.DeviceID == nil && (task.GroupID == nil || device.GroupID == nil || *task.GroupID != *device.GroupID) {
			continue
		}
		h.taskStore.SetTaskStatus(task.ID, "completed")
		writeLog(h.logStore, "info", "updates", "Task completed.", task.PackageName+" "+task.PackageVersion+" @ "+device.Name, task.CreatedBy)
	}
}
