package handlers

import (
	"bytes"
	"context"
	"errors"
	"goth/internal/agent"
	"goth/internal/middleware"
	"goth/internal/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type statusGetterMock struct {
	mock.Mock
}

func (m *statusGetterMock) GetStatus(baseURL string) (*agent.Status, error) {
	args := m.Called(baseURL)
	status, _ := args.Get(0).(*agent.Status)
	return status, args.Error(1)
}

func TestGetComputers(t *testing.T) {
	testCases := []struct {
		name               string
		user               *store.User
		agents             []string
		stub               func(m *statusGetterMock)
		expectedStatusCode int
		expectedLocation   string
		expectedBody       []string
	}{
		{
			name: "online and offline agents render",
			user: &store.User{ID: 1, Email: "test@example.com"},
			agents: []string{"http://localhost:4101", "http://localhost:4199"},
			stub: func(m *statusGetterMock) {
				m.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
					Name: "lab-pc-1", Version: "3.1.4", OS: "NixOS 25.05", UptimeSeconds: 3661,
				}, nil)
				m.On("GetStatus", "http://localhost:4199").Return(nil, errors.New("agent unreachable"))
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"lab-pc-1", "3.1.4", "NixOS 25.05", "1 ч 1 мин", "Офлайн", "—"},
		},
		{
			name:               "empty agent list shows empty state",
			user:               &store.User{ID: 1, Email: "test@example.com"},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"Няма конфигурирани компютри."},
		},
		{
			name:               "no user redirects to login",
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "/login",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			statusGetter := &statusGetterMock{}
			if tc.stub != nil {
				tc.stub(statusGetter)
			}

			handler := NewComputersHandler(ComputersHandlerParams{
				Agents:       tc.agents,
				StatusGetter: statusGetter,
			})

			req, _ := http.NewRequest("GET", "/computers", nil)
			ctx := context.WithValue(req.Context(), middleware.NonceKey, middleware.Nonces{
				Htmx:            "nonce-1234",
				ResponseTargets: "nonce-5678",
				Tw:              "nonce-9101",
			})
			if tc.user != nil {
				ctx = context.WithValue(ctx, middleware.UserKey, tc.user)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(tc.expectedStatusCode, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)

			if tc.expectedLocation != "" {
				assert.Equal(tc.expectedLocation, rr.Header().Get("Location"), "handler returned wrong redirect: got %v want %v", rr.Header().Get("Location"), tc.expectedLocation)
			}

			for _, expected := range tc.expectedBody {
				assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q, got %v", expected, rr.Body.String())
			}

			statusGetter.AssertExpectations(t)
		})
	}
}
