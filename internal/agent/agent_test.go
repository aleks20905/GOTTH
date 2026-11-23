package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatusHandler(t *testing.T) {
	assert := assert.New(t)

	handler := NewStatusHandler("test-pc", "2.0.1", "NixOS 25.05", "")
	handler.startedAt = time.Now().Add(-90 * time.Second)

	req, _ := http.NewRequest("GET", "/status", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)

	var status Status
	assert.NoError(json.NewDecoder(rr.Body).Decode(&status))
	assert.Equal("test-pc", status.Name)
	assert.Equal("2.0.1", status.Version)
	assert.Equal("NixOS 25.05", status.OS)
	assert.GreaterOrEqual(status.UptimeSeconds, int64(89), "uptime should be seconds since start, got %v", status.UptimeSeconds)
}

func TestStatusHandlerRequiresToken(t *testing.T) {
	assert := assert.New(t)

	handler := NewStatusHandler("test-pc", "2.0.1", "NixOS 25.05", "s3cret")
	handler.startedAt = time.Now()

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusUnauthorized, rr.Code, "open poll must be rejected")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/status", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusUnauthorized, rr.Code, "wrong token must be rejected")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/status", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusOK, rr.Code, "right token must pass")

	var status Status
	assert.NoError(json.NewDecoder(rr.Body).Decode(&status))
	assert.Equal("test-pc", status.Name)
}

func TestSendRegistration(t *testing.T) {
	assert := assert.New(t)

	var got Registration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("/api/agents/register", r.URL.Path)
		assert.NoError(json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := SendRegistration(server.URL, Registration{
		Name: "lab-pc-1", Version: "3.1.4", OS: "NixOS 25.05",
		UptimeSeconds: 60, URL: "http://localhost:4101", Token: "s3cret",
	})

	assert.NoError(err)
	assert.Equal("lab-pc-1", got.Name)
	assert.Equal("http://localhost:4101", got.URL)
	assert.Equal("s3cret", got.Token)
}

func TestSendRegistrationRejected(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	err := SendRegistration(server.URL, Registration{Token: "wrong"})

	assert.Error(err)
}

func TestFirewallHandler(t *testing.T) {
	assert := assert.New(t)

	handler := NewFirewallHandler(NewRuleStore(DefaultFirewallRules()), "s3cret")

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/firewall", nil)
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusUnauthorized, rr.Code, "open poll must be rejected")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/firewall", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusOK, rr.Code)

	var rules []FirewallRule
	assert.NoError(json.NewDecoder(rr.Body).Decode(&rules))
	assert.Len(rules, 2, "seeded defaults must be listed")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/firewall/apply", strings.NewReader(`{"rules":[{"direction":"inbound","protocol":"tcp","port_from":22,"port_to":22,"action":"allow","enabled":true}]}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusNoContent, rr.Code)

	assert.Len(handler.store.List(), 1, "apply must replace the rule set")
	assert.Equal(22, handler.store.List()[0].PortFrom)
}

func TestUpdateHandler(t *testing.T) {
	assert := assert.New(t)

	oldDelay := installDelay
	installDelay = 20 * time.Millisecond
	defer func() { installDelay = oldDelay }()

	status := NewStatusHandler("test-pc", "1.0.0", "NixOS 25.05", "s3cret")
	handler := NewUpdateHandler(status, "s3cret")

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/update", strings.NewReader(`{"package":"office-agent","version":"2.0.0"}`))
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusUnauthorized, rr.Code, "open install must be rejected")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/update", strings.NewReader(`{"package":"office-agent","version":"2.0.0"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusAccepted, rr.Code)

	deadline := time.Now().Add(2 * time.Second)
	for status.Snapshot().Version != "2.0.0" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal("2.0.0", status.Snapshot().Version, "version must bump after the fake install")
	assert.False(status.Snapshot().Updating, "updating flag must clear after install")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/update", strings.NewReader(`{"package":"office-agent"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusBadRequest, rr.Code, "missing version must be rejected")
}

func TestPowerHandler(t *testing.T) {
	assert := assert.New(t)

	status := NewStatusHandler("test-pc", "1.0.0", "NixOS 25.05", "s3cret")
	handler := NewPowerHandler(status, "s3cret")

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/power", strings.NewReader(`{"action":"reboot"}`))
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusUnauthorized, rr.Code, "open power call must be rejected")

	oldDelay := rebootDelay
	rebootDelay = 20 * time.Millisecond
	defer func() { rebootDelay = oldDelay }()

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/power", strings.NewReader(`{"action":"reboot"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusAccepted, rr.Code)
	assert.False(status.IsPowered(), "reboot must power off first")

	deadline := time.Now().Add(2 * time.Second)
	for !status.IsPowered() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	assert.True(status.IsPowered(), "reboot must power back on")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/power", strings.NewReader(`{"action":"shutdown"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusAccepted, rr.Code)
	assert.False(status.IsPowered(), "shutdown must stay off")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/status", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	status.ServeHTTP(rr, req)
	assert.Equal(http.StatusServiceUnavailable, rr.Code, "powered off agent must refuse status")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/power", strings.NewReader(`{"action":"wakeup"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusAccepted, rr.Code)
	assert.True(status.IsPowered(), "wakeup must power on")

	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/power", strings.NewReader(`{"action":"overclock"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusBadRequest, rr.Code, "unknown action must be rejected")
}

func TestConsoleHandler(t *testing.T) {
	assert := assert.New(t)

	status := NewStatusHandler("test-pc", "2.0.1", "NixOS 25.05", "s3cret")
	status.startedAt = time.Now().Add(-90 * time.Second)
	handler := NewConsoleHandler(status, "s3cret")

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/console", strings.NewReader(`{"command":"whoami"}`))
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusUnauthorized, rr.Code, "open console must be rejected")

	for command, expected := range map[string]string{
		"whoami":  "test-pc",
		"version": "2.0.1",
		"uptime":  "up ",
		"help":    "whoami",
		"rm -rf":  "unknown command",
	} {
		rr = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/console", strings.NewReader(`{"command":"`+command+`"}`))
		req.Header.Set("Authorization", "Bearer s3cret")
		handler.ServeHTTP(rr, req)
		assert.Equal(http.StatusOK, rr.Code)
		assert.Contains(rr.Body.String(), expected, "command %q", command)
	}

	status.SetPowered(false)
	rr = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/console", strings.NewReader(`{"command":"whoami"}`))
	req.Header.Set("Authorization", "Bearer s3cret")
	handler.ServeHTTP(rr, req)
	assert.Equal(http.StatusServiceUnavailable, rr.Code, "powered off agent must refuse console")
}

func TestClientGetStatus(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method)
		assert.Equal("/status", r.URL.Path)
		w.Write([]byte(`{"name":"lab-pc-1","version":"3.1.4","os":"NixOS 25.05","uptime_seconds":3661}`))
	}))
	defer server.Close()

	client := NewClient("")
	status, err := client.GetStatus(server.URL)

	assert.NoError(err)
	assert.Equal("lab-pc-1", status.Name)
	assert.Equal("3.1.4", status.Version)
	assert.Equal("NixOS 25.05", status.OS)
	assert.Equal(int64(3661), status.UptimeSeconds)
}

func TestClientGetStatusUnreachable(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := NewClient("")
	status, err := client.GetStatus(server.URL)

	assert.Error(err)
	assert.Nil(status)
}

func TestClientGetStatusTimesOut(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("")
	before := time.Now()
	status, err := client.GetStatus(server.URL)

	assert.Error(err)
	assert.Nil(status)
	assert.LessOrEqual(time.Since(before), 550*time.Millisecond, "client should give up after 500ms, took %v", time.Since(before))
}
