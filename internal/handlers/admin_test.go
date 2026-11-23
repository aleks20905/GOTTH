package handlers

import (
	"bytes"
	"context"
	"goth/internal/agent"
	"goth/internal/middleware"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func adminTestContext(user *store.User, isAdmin bool) context.Context {
	ctx := computersTestContext(user)
	if user != nil {
		ctx = context.WithValue(ctx, middleware.AdminKey, isAdmin)
	}
	return ctx
}

func TestAdminPage(t *testing.T) {
	testCases := []struct {
		name               string
		user               *store.User
		isAdmin            bool
		devices            []store.Device
		expectedStatusCode int
		expectedLocation   string
		expectedBody       []string
	}{
		{
			name:               "admin sees devices and add form",
			user:               &store.User{ID: 1, Email: "boss@example.com"},
			isAdmin:            true,
			devices:            []store.Device{{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"}},
			expectedStatusCode: http.StatusOK,
			expectedBody:       []string{"Administration", "lab-pc-1", "Online", "Add", "Delete"},
		},
		{
			name:               "no user redirects to login",
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "/login",
		},
		{
			name:               "non-admin redirects home",
			user:               &store.User{ID: 2, Email: "user@example.com"},
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			deviceStore := &storemock.DeviceStoreMock{}
			if tc.user != nil && tc.isAdmin {
				deviceStore.On("ListDevices").Return(tc.devices, nil)
			}
			statusGetter := &statusGetterMock{}
			if tc.user != nil && tc.isAdmin {
				statusGetter.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
					Name: "lab-pc-1", Version: "3.1.4", OS: "NixOS 25.05", UptimeSeconds: 60,
				}, nil)
			}

			handler := NewAdminHandler(AdminHandlerParams{
				DeviceStore:  deviceStore,
				StatusGetter: statusGetter,
				LogStore:     &storemock.LogStoreMock{},
			})

			req, _ := http.NewRequest("GET", "/admin", nil)
			req = req.WithContext(adminTestContext(tc.user, tc.isAdmin))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(tc.expectedStatusCode, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)

			if tc.expectedLocation != "" {
				assert.Equal(tc.expectedLocation, rr.Header().Get("Location"), "handler returned wrong redirect: got %v want %v", rr.Header().Get("Location"), tc.expectedLocation)
			}

			for _, expected := range tc.expectedBody {
				assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q, got %v", expected, rr.Body.String())
			}

			deviceStore.AssertExpectations(t)
		})
	}
}

func TestAdminAddDevice(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	t.Run("valid device redirects to admin", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		logStore := &storemock.LogStoreMock{}
		deviceStore.On("AddDevice", &store.Device{Name: "lab-pc-1", URL: "http://localhost:4101"}).
			Return(&store.Device{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"}, nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		handler := NewAdminHandler(AdminHandlerParams{DeviceStore: deviceStore, LogStore: logStore})

		form := url.Values{"name": {"lab-pc-1"}, "url": {"http://localhost:4101"}}
		req, _ := http.NewRequest("POST", "/admin/devices", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(adminTestContext(admin, true))

		rr := httptest.NewRecorder()
		handler.AddDevice(rr, req)

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Equal("/admin", rr.Header().Get("Location"))
		deviceStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("missing name re-renders with error", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		deviceStore.On("ListDevices").Return([]store.Device{}, nil)

		handler := NewAdminHandler(AdminHandlerParams{
			DeviceStore:  deviceStore,
			StatusGetter: &statusGetterMock{},
			LogStore:     &storemock.LogStoreMock{},
		})

		form := url.Values{"url": {"http://localhost:4101"}}
		req, _ := http.NewRequest("POST", "/admin/devices", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(adminTestContext(admin, true))

		rr := httptest.NewRecorder()
		handler.AddDevice(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		assert.True(bytes.Contains(rr.Body.Bytes(), []byte("required")))
		deviceStore.AssertExpectations(t)
	})

	t.Run("bad url re-renders with error", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		deviceStore.On("ListDevices").Return([]store.Device{}, nil)

		handler := NewAdminHandler(AdminHandlerParams{
			DeviceStore:  deviceStore,
			StatusGetter: &statusGetterMock{},
			LogStore:     &storemock.LogStoreMock{},
		})

		form := url.Values{"name": {"pc"}, "url": {"not a url"}}
		req, _ := http.NewRequest("POST", "/admin/devices", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(adminTestContext(admin, true))

		rr := httptest.NewRecorder()
		handler.AddDevice(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		assert.True(bytes.Contains(rr.Body.Bytes(), []byte("valid URL")))
		deviceStore.AssertExpectations(t)
	})

	t.Run("non-admin cannot add", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}

		handler := NewAdminHandler(AdminHandlerParams{DeviceStore: deviceStore, LogStore: &storemock.LogStoreMock{}})

		form := url.Values{"name": {"pc"}, "url": {"http://localhost:4101"}}
		req, _ := http.NewRequest("POST", "/admin/devices", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(adminTestContext(&store.User{ID: 2, Email: "user@example.com"}, false))

		rr := httptest.NewRecorder()
		handler.AddDevice(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/", rr.Header().Get("Location"))
		deviceStore.AssertExpectations(t)
	})
}

func TestAdminDeleteDevice(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	newDeleteRequest := func(user *store.User, isAdmin bool) *http.Request {
		req, _ := http.NewRequest("POST", "/admin/devices/delete/3", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "3")
		ctx := adminTestContext(user, isAdmin)
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		return req.WithContext(ctx)
	}

	t.Run("admin delete redirects to admin", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		logStore := &storemock.LogStoreMock{}
		deviceStore.On("DeleteDevice", uint(3)).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		handler := NewAdminHandler(AdminHandlerParams{DeviceStore: deviceStore, LogStore: logStore})

		rr := httptest.NewRecorder()
		handler.DeleteDevice(rr, newDeleteRequest(admin, true))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Equal("/admin", rr.Header().Get("Location"))
		deviceStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("non-admin cannot delete", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}

		handler := NewAdminHandler(AdminHandlerParams{DeviceStore: deviceStore, LogStore: &storemock.LogStoreMock{}})

		rr := httptest.NewRecorder()
		handler.DeleteDevice(rr, newDeleteRequest(&store.User{ID: 2, Email: "user@example.com"}, false))

		assert.Equal(http.StatusFound, rr.Code)
		deviceStore.AssertExpectations(t)
	})
}
