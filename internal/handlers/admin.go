package handlers

import (
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type AdminHandlerParams struct {
	DeviceStore  store.DeviceStore
	StatusGetter StatusGetter
	LogStore     store.LogStore
}

type AdminHandler struct {
	deviceStore  store.DeviceStore
	statusGetter StatusGetter
	logStore     store.LogStore
}

func NewAdminHandler(params AdminHandlerParams) *AdminHandler {
	return &AdminHandler{
		deviceStore:  params.DeviceStore,
		statusGetter: params.StatusGetter,
		logStore:     params.LogStore,
	}
}

func (h *AdminHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
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

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	h.render(w, r, "")
}

func (h *AdminHandler) render(w http.ResponseWriter, r *http.Request, errorMsg string) {
	devices, err := h.deviceStore.ListDevices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := templates.Admin(buildComputerRows(h.statusGetter, devices), errorMsg)
	if err := templates.Layout(c, "Administration").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *AdminHandler) AddDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	name := r.FormValue("name")
	deviceURL := r.FormValue("url")

	if name == "" || deviceURL == "" {
		h.render(w, r, "Name and address are required.")
		return
	}
	if _, err := url.ParseRequestURI(deviceURL); err != nil {
		h.render(w, r, "Address must be a valid URL, e.g. http://localhost:4101.")
		return
	}

	if _, err := h.deviceStore.AddDevice(&store.Device{Name: name, URL: deviceURL}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "devices", "Device added.", name+" "+deviceURL, email)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.deviceStore.DeleteDevice(uint(id)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "devices", "Device deleted.", strconv.FormatUint(id, 10), email)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
