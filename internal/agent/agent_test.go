package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatusHandler(t *testing.T) {
	assert := assert.New(t)

	handler := NewStatusHandler("test-pc", "2.0.1", "NixOS 25.05")
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

func TestClientGetStatus(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method)
		assert.Equal("/status", r.URL.Path)
		w.Write([]byte(`{"name":"lab-pc-1","version":"3.1.4","os":"NixOS 25.05","uptime_seconds":3661}`))
	}))
	defer server.Close()

	client := NewClient()
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

	client := NewClient()
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

	client := NewClient()
	before := time.Now()
	status, err := client.GetStatus(server.URL)

	assert.Error(err)
	assert.Nil(status)
	assert.LessOrEqual(time.Since(before), 550*time.Millisecond, "client should give up after 500ms, took %v", time.Since(before))
}
