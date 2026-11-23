package handlers

import (
	"bytes"
	"errors"
	"goth/internal/agent"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHome(t *testing.T) {
	t.Run("guest sees landing", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewHomeHandler(HomeDataParams{})

		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(computersTestContext(nil))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		assert.True(bytes.Contains(rr.Body.Bytes(), []byte("Welcome to")))
	})

	t.Run("user sees stats and activity", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}
		statusGetter := &statusGetterMock{}
		deviceStore.On("ListDevices").Return([]store.Device{
			{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"},
			{ID: 2, Name: "dead-pc", URL: "http://localhost:4199"},
		}, nil)
		statusGetter.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
			Name: "lab-pc-1", Version: "2.0.0", OS: "NixOS 25.05", UptimeSeconds: 60,
		}, nil)
		statusGetter.On("GetStatus", "http://localhost:4199").Return(nil, errors.New("down"))
		taskStore.On("ListTasksByStatus", "pending").Return([]store.UpdateTask{{ID: 1}}, nil)
		taskStore.On("ListTasksByStatus", "running").Return([]store.UpdateTask{}, nil)
		logStore.On("ListEntries", 5).Return([]store.LogEntry{
			{ID: 1, Level: "info", Message: "Device added."},
		}, nil)

		handler := NewHomeHandler(HomeDataParams{
			DeviceStore:  deviceStore,
			TaskStore:    taskStore,
			LogStore:     logStore,
			StatusGetter: statusGetter,
		})

		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(computersTestContext(&store.User{ID: 1, Email: "me@example.com"}))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		for _, expected := range []string{"Recent activity", "Device added.", "Online", "Open tasks"} {
			assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q", expected)
		}
		deviceStore.AssertExpectations(t)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
		statusGetter.AssertExpectations(t)
	})

	t.Run("store error is a 500", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		deviceStore.On("ListDevices").Return(nil, errors.New("db down"))

		handler := NewHomeHandler(HomeDataParams{DeviceStore: deviceStore})

		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(computersTestContext(&store.User{ID: 1, Email: "me@example.com"}))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusInternalServerError, rr.Code)
		deviceStore.AssertExpectations(t)
	})
}
