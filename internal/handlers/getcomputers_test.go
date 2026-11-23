package handlers

import (
	"bytes"
	"context"
	"errors"
	"goth/internal/agent"
	"goth/internal/middleware"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func computersTestContext(user *store.User) context.Context {
	ctx := context.WithValue(context.Background(), middleware.NonceKey, middleware.Nonces{
		Htmx:            "nonce-1234",
		ResponseTargets: "nonce-5678",
		Tw:              "nonce-9101",
	})
	if user != nil {
		ctx = context.WithValue(ctx, middleware.UserKey, user)
	}
	return ctx
}

func TestGetComputers(t *testing.T) {
	groupID := uint(1)
	testCases := []struct {
		name               string
		user               *store.User
		query              string
		devices            []store.Device
		groups             []store.DeviceGroup
		listErr            error
		stub               func(m *statusGetterMock)
		expectedStatusCode int
		expectedLocation   string
		expectedBody       []string
		unexpectedBody     []string
	}{
		{
			name: "online and offline devices render",
			user: &store.User{ID: 1, Email: "test@example.com"},
			devices: []store.Device{
				{ID: 1, Name: "stored-pc-1", URL: "http://localhost:4101"},
				{ID: 2, Name: "stored-pc-2", URL: "http://localhost:4199"},
			},
			stub: func(m *statusGetterMock) {
				m.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
					Name: "lab-pc-1", Version: "3.1.4", OS: "NixOS 25.05", UptimeSeconds: 3661,
				}, nil)
				m.On("GetStatus", "http://localhost:4199").Return(nil, errors.New("agent unreachable"))
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"lab-pc-1", "3.1.4", "NixOS 25.05", "1 h 1 min", "stored-pc-2", "Offline", "—"},
		},
		{
			name: "offline device shows last known version and last seen",
			user: &store.User{ID: 1, Email: "test@example.com"},
			devices: []store.Device{
				{ID: 1, Name: "stored-pc", URL: "http://localhost:4199", Version: "1.2.3", OS: "NixOS 25.05", Verified: true, LastSeen: time.Now().Add(-90 * time.Second)},
			},
			stub: func(m *statusGetterMock) {
				m.On("GetStatus", "http://localhost:4199").Return(nil, errors.New("agent unreachable"))
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"stored-pc", "1.2.3", "NixOS 25.05", "1 min ago", "Offline"},
		},
		{
			name: "updating device shows updating badge",
			user: &store.User{ID: 1, Email: "test@example.com"},
			devices: []store.Device{
				{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"},
			},
			stub: func(m *statusGetterMock) {
				m.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
					Name: "lab-pc-1", Version: "3.1.4", OS: "NixOS 25.05", UptimeSeconds: 60, Updating: true,
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"Updating"},
		},
		{
			name:  "group filter shows only members",
			user:  &store.User{ID: 1, Email: "test@example.com"},
			query: "?group=1",
			devices: []store.Device{
				{ID: 1, Name: "finance-pc", URL: "http://localhost:4101", GroupID: &groupID},
				{ID: 2, Name: "other-pc", URL: "http://localhost:4102"},
			},
			groups: []store.DeviceGroup{{ID: 1, Name: "Finance"}},
			stub: func(m *statusGetterMock) {
				m.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
					Name: "finance-pc", Version: "1.0.0", OS: "NixOS 25.05", UptimeSeconds: 60,
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"finance-pc", "Finance"},
			unexpectedBody:     []string{"other-pc"},
		},
		{
			name:               "empty device list shows empty state",
			user:               &store.User{ID: 1, Email: "test@example.com"},
			devices:            []store.Device{},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"No computers added."},
		},
		{
			name:               "store error is a 500",
			user:               &store.User{ID: 1, Email: "test@example.com"},
			listErr:            errors.New("db down"),
			expectedStatusCode: http.StatusInternalServerError,
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
			deviceStore := &storemock.DeviceStoreMock{}
			if tc.user != nil {
				deviceStore.On("ListDevices").Return(tc.devices, tc.listErr)
			}
			groupStore := &storemock.GroupStoreMock{}
			if tc.user != nil && tc.listErr == nil {
				groupStore.On("ListGroups").Return(tc.groups, nil)
			}
			statusGetter := &statusGetterMock{}
			if tc.stub != nil {
				tc.stub(statusGetter)
			}

			handler := NewComputersHandler(ComputersHandlerParams{
				DeviceStore:  deviceStore,
				GroupStore:   groupStore,
				StatusGetter: statusGetter,
			})

			req, _ := http.NewRequest("GET", "/computers"+tc.query, nil)
			req = req.WithContext(computersTestContext(tc.user))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(tc.expectedStatusCode, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)

			if tc.expectedLocation != "" {
				assert.Equal(tc.expectedLocation, rr.Header().Get("Location"), "handler returned wrong redirect: got %v want %v", rr.Header().Get("Location"), tc.expectedLocation)
			}

			for _, expected := range tc.expectedBody {
				assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q, got %v", expected, rr.Body.String())
			}
			for _, unexpected := range tc.unexpectedBody {
				assert.False(bytes.Contains(rr.Body.Bytes(), []byte(unexpected)), "body should not contain %q", unexpected)
			}

			deviceStore.AssertExpectations(t)
			groupStore.AssertExpectations(t)
			statusGetter.AssertExpectations(t)
		})
	}
}
