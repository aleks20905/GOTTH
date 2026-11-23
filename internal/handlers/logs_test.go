package handlers

import (
	"bytes"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogsPage(t *testing.T) {
	t.Run("page renders entries", func(t *testing.T) {
		assert := assert.New(t)
		logStore := &storemock.LogStoreMock{}
		logStore.On("ListEntries", 200).Return([]store.LogEntry{
			{ID: 1, Level: "info", Category: "devices", Message: "Device added.", UserEmail: "boss@example.com"},
			{ID: 2, Level: "error", Category: "updates", Message: "Task failed.", UserEmail: "boss@example.com"},
		}, nil)

		handler := NewLogsHandler(LogsHandlerParams{LogStore: logStore})

		req, _ := http.NewRequest("GET", "/logs", nil)
		req = req.WithContext(computersTestContext(&store.User{ID: 1, Email: "boss@example.com"}))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		for _, expected := range []string{"Logs", "Device added.", "Task failed.", "error"} {
			assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q", expected)
		}
		logStore.AssertExpectations(t)
	})

	t.Run("anonymous user redirects to login", func(t *testing.T) {
		assert := assert.New(t)
		logStore := &storemock.LogStoreMock{}

		handler := NewLogsHandler(LogsHandlerParams{LogStore: logStore})

		req, _ := http.NewRequest("GET", "/logs", nil)
		req = req.WithContext(computersTestContext(nil))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/login", rr.Header().Get("Location"))
		logStore.AssertExpectations(t)
	})
}
