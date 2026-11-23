package handlers

import (
	"goth/internal/agent"
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type DeviceAgentClient interface {
	ApplyFirewall(baseURL string, rules []agent.FirewallRule) error
	SendPower(baseURL string, action string) error
	SendCommand(baseURL string, command string) (string, error)
}

type DeviceHandlerParams struct {
	DeviceStore   store.DeviceStore
	FirewallStore store.FirewallStore
	TaskStore     store.TaskStore
	LogStore      store.LogStore
	StatusGetter  StatusGetter
	Updater       DeviceAgentClient
}

type DeviceHandler struct {
	deviceStore   store.DeviceStore
	firewallStore store.FirewallStore
	taskStore     store.TaskStore
	logStore      store.LogStore
	statusGetter  StatusGetter
	updater       DeviceAgentClient
}

func NewDeviceHandler(params DeviceHandlerParams) *DeviceHandler {
	return &DeviceHandler{
		deviceStore:   params.DeviceStore,
		firewallStore: params.FirewallStore,
		taskStore:     params.TaskStore,
		logStore:      params.LogStore,
		statusGetter:  params.StatusGetter,
		updater:       params.Updater,
	}
}

func (h *DeviceHandler) loadDevice(w http.ResponseWriter, r *http.Request) (*store.Device, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, false
	}

	device, err := h.deviceStore.GetDevice(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, false
	}
	return device, true
}

func (h *DeviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	device, ok := h.loadDevice(w, r)
	if !ok {
		return
	}

	rules, err := h.firewallStore.ListRules(device.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks, err := h.taskStore.ListTasksForDevice(device.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	row := buildComputerRows(h.statusGetter, []store.Device{*device})[0]

	c := templates.DeviceDetail(*device, row, rules, tasks, r.URL.Query().Get("applied"), r.URL.Query().Get("power"), "")
	if err := templates.Layout(c, device.Name).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *DeviceHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
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

func (h *DeviceHandler) deviceRedirect(w http.ResponseWriter, r *http.Request, device *store.Device, applied string) {
	target := "/devices/" + strconv.FormatUint(uint64(device.ID), 10)
	if applied != "" {
		target += "?applied=" + applied
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

var validDirections = map[string]bool{"inbound": true, "outbound": true}
var validProtocols = map[string]bool{"tcp": true, "udp": true}
var validActions = map[string]bool{"allow": true, "deny": true}

func (h *DeviceHandler) AddRule(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	device, ok := h.loadDevice(w, r)
	if !ok {
		return
	}

	rule, valid := parseRuleForm(r)
	if !valid {
		h.deviceRedirect(w, r, device, "invalid")
		return
	}
	rule.DeviceID = device.ID

	if _, err := h.firewallStore.AddRule(&rule); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "firewall", "Rule added.", device.Name, email)
	h.deviceRedirect(w, r, device, "")
}

func parseRuleForm(r *http.Request) (store.FirewallRule, bool) {
	from, err1 := strconv.Atoi(r.FormValue("port_from"))
	to, err2 := strconv.Atoi(r.FormValue("port_to"))
	rule := store.FirewallRule{
		Direction:     r.FormValue("direction"),
		Protocol:      r.FormValue("protocol"),
		PortFrom:      from,
		PortTo:        to,
		RemoteAddress: r.FormValue("remote_address"),
		Action:        r.FormValue("action"),
		Enabled:       true,
	}

	if !validDirections[rule.Direction] || !validProtocols[rule.Protocol] || !validActions[rule.Action] {
		return rule, false
	}
	if err1 != nil || err2 != nil || from < 1 || to > 65535 || from > to {
		return rule, false
	}
	return rule, true
}

func (h *DeviceHandler) ruleRedirectTarget(w http.ResponseWriter, r *http.Request) (*store.Device, uint64, bool) {
	device, ok := h.loadDevice(w, r)
	if !ok {
		return nil, 0, false
	}
	ruleID, err := strconv.ParseUint(chi.URLParam(r, "ruleID"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, 0, false
	}
	return device, ruleID, true
}

func (h *DeviceHandler) ToggleRule(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	device, ruleID, ok := h.ruleRedirectTarget(w, r)
	if !ok {
		return
	}

	enabled := r.FormValue("enabled") == "1"
	if err := h.firewallStore.SetRuleEnabled(uint(ruleID), enabled); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "firewall", "Rule toggled.", device.Name, email)
	h.deviceRedirect(w, r, device, "")
}

func (h *DeviceHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	device, ruleID, ok := h.ruleRedirectTarget(w, r)
	if !ok {
		return
	}

	if err := h.firewallStore.DeleteRule(uint(ruleID)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeLog(h.logStore, "info", "firewall", "Rule deleted.", device.Name, email)
	h.deviceRedirect(w, r, device, "")
}

func (h *DeviceHandler) ApplyRules(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	device, ok := h.loadDevice(w, r)
	if !ok {
		return
	}

	rules, err := h.firewallStore.ListRules(device.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload := make([]agent.FirewallRule, 0, len(rules))
	for _, rule := range rules {
		payload = append(payload, agent.FirewallRule{
			Direction:     rule.Direction,
			Protocol:      rule.Protocol,
			PortFrom:      rule.PortFrom,
			PortTo:        rule.PortTo,
			RemoteAddress: rule.RemoteAddress,
			Action:        rule.Action,
			Enabled:       rule.Enabled,
		})
	}

	if err := h.updater.ApplyFirewall(device.URL, payload); err != nil {
		writeLog(h.logStore, "error", "firewall", "Apply failed.", device.Name, email)
		h.deviceRedirect(w, r, device, "fail")
		return
	}

	writeLog(h.logStore, "info", "firewall", "Rules applied.", device.Name, email)
	h.deviceRedirect(w, r, device, "ok")
}

func (h *DeviceHandler) PowerDevice(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	device, ok := h.loadDevice(w, r)
	if !ok {
		return
	}

	action := r.FormValue("action")
	if action != "reboot" && action != "shutdown" && action != "wakeup" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	target := "/devices/" + strconv.FormatUint(uint64(device.ID), 10)
	if err := h.updater.SendPower(device.URL, action); err != nil {
		writeLog(h.logStore, "error", "power", "Power command failed.", device.Name, email)
		http.Redirect(w, r, target+"?power=fail", http.StatusSeeOther)
		return
	}

	writeLog(h.logStore, "info", "power", "Power command sent: "+action+".", device.Name, email)
	http.Redirect(w, r, target+"?power=ok", http.StatusSeeOther)
}

func (h *DeviceHandler) RunConsole(w http.ResponseWriter, r *http.Request) {
	email, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	device, ok := h.loadDevice(w, r)
	if !ok {
		return
	}

	command := r.FormValue("command")
	if command == "" {
		h.deviceRedirect(w, r, device, "")
		return
	}

	output, err := h.updater.SendCommand(device.URL, command)
	if err != nil {
		writeLog(h.logStore, "error", "console", "Console command failed.", device.Name, email)
		output = "error: device did not respond"
	}

	rules, err := h.firewallStore.ListRules(device.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks, err := h.taskStore.ListTasksForDevice(device.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	row := buildComputerRows(h.statusGetter, []store.Device{*device})[0]

	c := templates.DeviceDetail(*device, row, rules, tasks, "", "", "$ "+command+"\n"+output)
	if err := templates.Layout(c, device.Name).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
