package handlers

import (
	"bytes"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterAgent(t *testing.T) {
	newRequest := func(body string) *http.Request {
		req, _ := http.NewRequest("POST", "/api/agents/register", strings.NewReader(body))
		return req
	}

	newHandler := func(deviceStore *storemock.DeviceStoreMock, taskStore *storemock.TaskStoreMock, logStore *storemock.LogStoreMock, token string) *RegisterHandler {
		return NewRegisterHandler(RegisterHandlerParams{
			DeviceStore: deviceStore,
			TaskStore:   taskStore,
			LogStore:    logStore,
			AgentToken:  token,
		})
	}

	t.Run("valid token upserts verified device", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}
		deviceStore.On("UpsertDevice", mock.MatchedBy(func(d *store.Device) bool {
			return d.URL == "http://localhost:4101" && d.Name == "lab-pc-1" && d.Version == "3.1.4" && d.OS == "NixOS 25.05" && d.Verified && !d.LastSeen.IsZero()
		})).Return(&store.Device{ID: 1}, nil)
		taskStore.On("ListTasksByStatus", "running").Return([]store.UpdateTask{}, nil)

		handler := newHandler(deviceStore, taskStore, logStore, "s3cret")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRequest(`{"name":"lab-pc-1","version":"3.1.4","os":"NixOS 25.05","uptime_seconds":60,"url":"http://localhost:4101","token":"s3cret"}`))

		assert.Equal(http.StatusNoContent, rr.Code)
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("heartbeat completes matching running task", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}
		deviceID := uint(1)
		deviceStore.On("UpsertDevice", mock.Anything).Return(&store.Device{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"}, nil)
		taskStore.On("ListTasksByStatus", "running").Return([]store.UpdateTask{
			{ID: 7, PackageName: "office-agent", PackageVersion: "3.1.4", DeviceID: &deviceID, Status: "running", CreatedBy: "boss@example.com"},
			{ID: 8, PackageName: "office-agent", PackageVersion: "9.9.9", DeviceID: &deviceID, Status: "running", CreatedBy: "boss@example.com"},
		}, nil)
		taskStore.On("SetTaskStatus", uint(7), "completed").Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		handler := newHandler(deviceStore, taskStore, logStore, "s3cret")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRequest(`{"name":"lab-pc-1","version":"3.1.4","os":"NixOS 25.05","uptime_seconds":60,"url":"http://localhost:4101","token":"s3cret"}`))

		assert.Equal(http.StatusNoContent, rr.Code)
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("wrong token is rejected", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}

		handler := newHandler(deviceStore, taskStore, logStore, "s3cret")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRequest(`{"name":"rogue","url":"http://evil:1","token":"nope"}`))

		assert.Equal(http.StatusUnauthorized, rr.Code)
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("empty server token rejects everything", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}

		handler := newHandler(deviceStore, taskStore, logStore, "")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRequest(`{"name":"pc","url":"http://localhost:4101","token":"s3cret"}`))

		assert.Equal(http.StatusUnauthorized, rr.Code)
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("bad json is rejected", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}

		handler := newHandler(deviceStore, taskStore, logStore, "s3cret")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newRequest(`{oops`))

		assert.Equal(http.StatusBadRequest, rr.Code)
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("non-post is rejected", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}

		handler := newHandler(deviceStore, taskStore, logStore, "s3cret")

		req, _ := http.NewRequest("GET", "/api/agents/register", bytes.NewReader(nil))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusMethodNotAllowed, rr.Code)
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})
}
