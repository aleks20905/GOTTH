package handlers

import (
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GroupsHandlerParams struct {
	DeviceStore store.DeviceStore
	GroupStore  store.GroupStore
	LogStore    store.LogStore
}

type GroupsHandler struct {
	deviceStore store.DeviceStore
	groupStore  store.GroupStore
	logStore    store.LogStore
}

func NewGroupsHandler(params GroupsHandlerParams) *GroupsHandler {
	return &GroupsHandler{
		deviceStore: params.DeviceStore,
		groupStore:  params.GroupStore,
		logStore:    params.LogStore,
	}
}

func (h *GroupsHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
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

func (h *GroupsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}

	groups, err := h.groupStore.ListGroups()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	devices, err := h.deviceStore.ListDevices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := templates.Groups(groups, devices, r.URL.Query().Get("error"))
	if err := templates.Layout(c, "Groups").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *GroupsHandler) AddGroup(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Redirect(w, r, "/admin/groups?error=empty", http.StatusSeeOther)
		return
	}

	if _, err := h.groupStore.AddGroup(&store.DeviceGroup{Name: name, Description: r.FormValue("description")}); err != nil {
		http.Redirect(w, r, "/admin/groups?error=exists", http.StatusSeeOther)
		return
	}

	writeLog(h.logStore, "info", "groups", "Group created.", name, email)
	http.Redirect(w, r, "/admin/groups", http.StatusSeeOther)
}

func (h *GroupsHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.groupStore.DeleteGroup(uint(id)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "groups", "Group deleted.", strconv.FormatUint(id, 10), email)
	http.Redirect(w, r, "/admin/groups", http.StatusSeeOther)
}

func (h *GroupsHandler) AssignDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	groupID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deviceID, err := strconv.ParseUint(r.FormValue("device_id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	gid := uint(groupID)
	if err := h.deviceStore.AssignGroup(uint(deviceID), &gid); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "groups", "Device added to group.", strconv.FormatUint(deviceID, 10), email)
	http.Redirect(w, r, "/admin/groups", http.StatusSeeOther)
}

func (h *GroupsHandler) UnassignDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	deviceID, err := strconv.ParseUint(chi.URLParam(r, "deviceID"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.deviceStore.AssignGroup(uint(deviceID), nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "groups", "Device removed from group.", strconv.FormatUint(deviceID, 10), email)
	http.Redirect(w, r, "/admin/groups", http.StatusSeeOther)
}
