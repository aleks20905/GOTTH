package handlers

import (
	"goth/internal/middleware"
	"goth/internal/scheduler"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"strconv"
	"time"
)

type UpdatesHandlerParams struct {
	DeviceStore store.DeviceStore
	GroupStore  store.GroupStore
	TaskStore   store.TaskStore
	LogStore    store.LogStore
	Updater     scheduler.Updater
}

type UpdatesHandler struct {
	deviceStore store.DeviceStore
	groupStore  store.GroupStore
	taskStore   store.TaskStore
	logStore    store.LogStore
	updater     scheduler.Updater
}

func NewUpdatesHandler(params UpdatesHandlerParams) *UpdatesHandler {
	return &UpdatesHandler{
		deviceStore: params.DeviceStore,
		groupStore:  params.GroupStore,
		taskStore:   params.TaskStore,
		logStore:    params.LogStore,
		updater:     params.Updater,
	}
}

func (h *UpdatesHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return "", false
	}
	if !middleware.IsAdmin(r.Context()) {
		http.Redirect(w, r, "/", http.StatusFound)
		return "", false
	}
	return user.Email, true
}

func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}

	tasks, err := h.taskStore.ListTasks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	rows := make([]templates.TaskRow, 0, len(tasks))
	for _, task := range tasks {
		rows = append(rows, templates.TaskRow{Task: task, Target: h.taskTarget(task, devices, groups)})
	}

	c := templates.Updates(rows, devices, groups, r.URL.Query().Get("error"))
	if err := templates.Layout(c, "Updates").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *UpdatesHandler) taskTarget(task store.UpdateTask, devices []store.Device, groups []store.DeviceGroup) string {
	if task.DeviceID != nil {
		for _, device := range devices {
			if device.ID == *task.DeviceID {
				return device.Name
			}
		}
		return "—"
	}
	if task.GroupID != nil {
		for _, group := range groups {
			if group.ID == *task.GroupID {
				return "Group: " + group.Name
			}
		}
		return "—"
	}
	return "—"
}

func (h *UpdatesHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	pkg := r.FormValue("package")
	version := r.FormValue("version")
	if pkg == "" || version == "" {
		http.Redirect(w, r, "/admin/updates?error=empty", http.StatusSeeOther)
		return
	}

	task := &store.UpdateTask{
		PackageName:    pkg,
		PackageVersion: version,
		Status:         "pending",
		ScheduledAt:    time.Now(),
		CreatedBy:      email,
	}

	if deviceID, err := strconv.ParseUint(r.FormValue("device_id"), 10, 32); err == nil && deviceID != 0 {
		id := uint(deviceID)
		task.DeviceID = &id
	} else if groupID, err := strconv.ParseUint(r.FormValue("group_id"), 10, 32); err == nil && groupID != 0 {
		id := uint(groupID)
		task.GroupID = &id
	} else {
		http.Redirect(w, r, "/admin/updates?error=target", http.StatusSeeOther)
		return
	}

	if when := r.FormValue("scheduled_at"); when != "" {
		scheduled, err := time.Parse("2006-01-02T15:04", when)
		if err != nil {
			http.Redirect(w, r, "/admin/updates?error=time", http.StatusSeeOther)
			return
		}
		task.ScheduledAt = scheduled
	}

	created, err := h.taskStore.AddTask(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	task.ID = created.ID

	writeLog(h.logStore, "info", "updates", "Task created.", pkg+" "+version, email)

	if !task.ScheduledAt.After(time.Now()) {
		scheduler.Dispatch(h.deviceStore, h.taskStore, h.logStore, h.updater, task)
	}

	http.Redirect(w, r, "/admin/updates", http.StatusSeeOther)
}
