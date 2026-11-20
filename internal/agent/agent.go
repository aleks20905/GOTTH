package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Status is the JSON payload every agent serves on /status.
type Status struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	OS            string `json:"os"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type StatusHandler struct {
	name      string
	version   string
	os        string
	startedAt time.Time
}

func NewStatusHandler(name string, version string, operatingSystem string) *StatusHandler {
	return &StatusHandler{
		name:      name,
		version:   version,
		os:        operatingSystem,
		startedAt: time.Now(),
	}
}

func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Status{
		Name:          h.name,
		Version:       h.version,
		OS:            h.os,
		UptimeSeconds: int64(time.Since(h.startedAt).Seconds()),
	})
}

const statusTimeout = 500 * time.Millisecond

// Client fetches one agent's status. It satisfies handlers.StatusGetter.
type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: statusTimeout}}
}

func (c *Client) GetStatus(baseURL string) (*Status, error) {
	resp, err := c.httpClient.Get(baseURL + "/status")
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
