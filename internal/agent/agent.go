package agent

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Status is the JSON payload every agent serves on /status.
type Status struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	OS            string `json:"os"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	Updating      bool   `json:"updating"`
}

// Registration is what an agent POSTs to the web app to pair itself.
type Registration struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	OS            string `json:"os"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	URL           string `json:"url"`
	Token         string `json:"token"`
}

// FirewallRule is one emulated firewall rule on the agent.
type FirewallRule struct {
	Direction     string `json:"direction"`
	Protocol      string `json:"protocol"`
	PortFrom      int    `json:"port_from"`
	PortTo        int    `json:"port_to"`
	RemoteAddress string `json:"remote_address"`
	Action        string `json:"action"`
	Enabled       bool   `json:"enabled"`
}

func DefaultFirewallRules() []FirewallRule {
	return []FirewallRule{
		{Direction: "inbound", Protocol: "tcp", PortFrom: 3389, PortTo: 3389, Action: "allow", Enabled: true},
		{Direction: "outbound", Protocol: "tcp", PortFrom: 1, PortTo: 65535, Action: "allow", Enabled: true},
	}
}

// TokensMatch compares secrets in constant time. Both sides use it.
func TokensMatch(secret, candidate string) bool {
	expected := sha256.Sum256([]byte(secret))
	got := sha256.Sum256([]byte(candidate))
	return subtle.ConstantTimeCompare(expected[:], got[:]) == 1
}

type StatusHandler struct {
	mu        sync.RWMutex
	name      string
	version   string
	os        string
	startedAt time.Time
	updating  bool
	powered   bool
	token     string
}

func NewStatusHandler(name string, version string, operatingSystem string, token string) *StatusHandler {
	return &StatusHandler{
		name:      name,
		version:   version,
		os:        operatingSystem,
		startedAt: time.Now(),
		powered:   true,
		token:     token,
	}
}

func (h *StatusHandler) Snapshot() Status {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return Status{
		Name:          h.name,
		Version:       h.version,
		OS:            h.os,
		UptimeSeconds: int64(time.Since(h.startedAt).Seconds()),
		Updating:      h.updating,
	}
}

func (h *StatusHandler) SetVersion(version string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.version = version
}

func (h *StatusHandler) SetUpdating(updating bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.updating = updating
}

func (h *StatusHandler) IsPowered() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.powered
}

func (h *StatusHandler) SetPowered(powered bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.powered = powered
}

func (h *StatusHandler) authorized(r *http.Request) bool {
	return h.token == "" || TokensMatch(h.token, bearerToken(r))
}

func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !h.IsPowered() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Snapshot())
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		return ""
	}
	return auth[len(prefix):]
}

// RuleStore is the agent's emulated firewall: plain memory, no OS calls.
type RuleStore struct {
	mu    sync.RWMutex
	rules []FirewallRule
}

func NewRuleStore(rules []FirewallRule) *RuleStore {
	return &RuleStore{rules: rules}
}

func (s *RuleStore) List() []FirewallRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FirewallRule, len(s.rules))
	copy(out, s.rules)
	return out
}

func (s *RuleStore) Replace(rules []FirewallRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = rules
}

type FirewallHandler struct {
	store *RuleStore
	token string
}

func NewFirewallHandler(store *RuleStore, token string) *FirewallHandler {
	return &FirewallHandler{store: store, token: token}
}

func (h *FirewallHandler) authorized(r *http.Request) bool {
	return h.token == "" || TokensMatch(h.token, bearerToken(r))
}

func (h *FirewallHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/firewall":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.store.List())
	case r.Method == http.MethodPost && r.URL.Path == "/firewall/apply":
		var payload struct {
			Rules []FirewallRule `json:"rules"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if payload.Rules == nil {
			payload.Rules = []FirewallRule{}
		}
		h.store.Replace(payload.Rules)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

type UpdateRequest struct {
	Package string `json:"package"`
	Version string `json:"version"`
}

var installDelay = 8 * time.Second

// UpdateHandler fakes an install: flips the updating flag, waits, bumps the version.
type UpdateHandler struct {
	status *StatusHandler
	token  string
}

func NewUpdateHandler(status *StatusHandler, token string) *UpdateHandler {
	return &UpdateHandler{status: status, token: token}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.token != "" && !TokensMatch(h.token, bearerToken(r)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost || r.URL.Path != "/update" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Version == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go h.install(req.Version)
	w.WriteHeader(http.StatusAccepted)
}

func (h *UpdateHandler) install(version string) {
	h.status.SetUpdating(true)
	defer h.status.SetUpdating(false)
	time.Sleep(installDelay)
	h.status.SetVersion(version)
}

var rebootDelay = 10 * time.Second

type PowerRequest struct {
	Action string `json:"action"`
}

// PowerHandler fakes power control: reboot powers off for a while, shutdown
// stays off until wakeup. While off, /status answers 503 and heartbeats stop.
type PowerHandler struct {
	status *StatusHandler
	token  string
}

func NewPowerHandler(status *StatusHandler, token string) *PowerHandler {
	return &PowerHandler{status: status, token: token}
}

func (h *PowerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.token != "" && !TokensMatch(h.token, bearerToken(r)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost || r.URL.Path != "/power" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var req PowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "reboot":
		h.status.SetPowered(false)
		go func() {
			time.Sleep(rebootDelay)
			h.status.SetPowered(true)
		}()
	case "shutdown":
		h.status.SetPowered(false)
	case "wakeup":
		h.status.SetPowered(true)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type ConsoleRequest struct {
	Command string `json:"command"`
}

type ConsoleResponse struct {
	Output string `json:"output"`
}

// ConsoleHandler fakes a remote shell with canned replies.
type ConsoleHandler struct {
	status *StatusHandler
	token  string
}

func NewConsoleHandler(status *StatusHandler, token string) *ConsoleHandler {
	return &ConsoleHandler{status: status, token: token}
}

func (h *ConsoleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.token != "" && !TokensMatch(h.token, bearerToken(r)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost || r.URL.Path != "/console" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !h.status.IsPowered() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var req ConsoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	snapshot := h.status.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ConsoleResponse{Output: cannedReply(snapshot, req.Command)})
}

func cannedReply(snapshot Status, command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}

	switch fields[0] {
	case "help":
		return "commands: help, whoami, version, uptime"
	case "whoami":
		return snapshot.Name
	case "version":
		return snapshot.Version
	case "uptime":
		return fmt.Sprintf("up %d seconds", snapshot.UptimeSeconds)
	default:
		return fmt.Sprintf("unknown command: %s (try help)", fields[0])
	}
}

const statusTimeout = 500 * time.Millisecond

// Client fetches one agent's status. It satisfies handlers.StatusGetter.
type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{httpClient: &http.Client{Timeout: statusTimeout}, token: token}
}

func (c *Client) doJSON(method string, url string, payload any, wantStatus int) error {
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		return fmt.Errorf("%s %s returned status %d", method, url, resp.StatusCode)
	}
	return nil
}

func (c *Client) GetStatus(baseURL string) (*Status, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/status", nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent %s returned status %d", baseURL, resp.StatusCode)
	}

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *Client) ApplyFirewall(baseURL string, rules []FirewallRule) error {
	return c.doJSON(http.MethodPost, baseURL+"/firewall/apply", map[string]any{"rules": rules}, http.StatusNoContent)
}

func (c *Client) StartUpdate(baseURL string, pkg string, version string) error {
	return c.doJSON(http.MethodPost, baseURL+"/update", UpdateRequest{Package: pkg, Version: version}, http.StatusAccepted)
}

func (c *Client) SendPower(baseURL string, action string) error {
	return c.doJSON(http.MethodPost, baseURL+"/power", PowerRequest{Action: action}, http.StatusAccepted)
}

func (c *Client) SendCommand(baseURL string, command string) (string, error) {
	payload, err := json.Marshal(ConsoleRequest{Command: command})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/console", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("POST %s/console returned status %d", baseURL, resp.StatusCode)
	}

	var out ConsoleResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Output, nil
}

const registerTimeout = 5 * time.Second

// SendRegistration phones one heartbeat home to the web app.
func SendRegistration(serverURL string, reg Registration) error {
	body, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: registerTimeout}
	resp, err := client.Post(serverURL+"/api/agents/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server %s returned status %d", serverURL, resp.StatusCode)
	}
	return nil
}
