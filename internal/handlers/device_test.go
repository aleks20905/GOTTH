package handlers

import (
	"bytes"
	"context"
	"errors"
	"goth/internal/agent"
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

type applyFirewallMock struct {
	mock.Mock
}

func (m *applyFirewallMock) ApplyFirewall(baseURL string, rules []agent.FirewallRule) error {
	args := m.Called(baseURL, rules)
	return args.Error(0)
}

func (m *applyFirewallMock) SendPower(baseURL string, action string) error {
	args := m.Called(baseURL, action)
	return args.Error(0)
}

func (m *applyFirewallMock) SendCommand(baseURL string, command string) (string, error) {
	args := m.Called(baseURL, command)
	out, _ := args.Get(0).(string)
	return out, args.Error(1)
}

func deviceRequest(user *store.User, isAdmin bool, method string, target string, params map[string]string, form url.Values) *http.Request {
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	req, _ := http.NewRequest(method, target, body)
	if form != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	ctx := adminTestContext(user, isAdmin)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestDevicePage(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}
	device := &store.Device{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101", Verified: true}

	newHandler := func() (*DeviceHandler, *storemock.DeviceStoreMock, *storemock.FirewallStoreMock, *storemock.TaskStoreMock, *statusGetterMock, *applyFirewallMock) {
		deviceStore := &storemock.DeviceStoreMock{}
		firewallStore := &storemock.FirewallStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		statusGetter := &statusGetterMock{}
		updater := &applyFirewallMock{}
		handler := NewDeviceHandler(DeviceHandlerParams{
			DeviceStore:   deviceStore,
			FirewallStore: firewallStore,
			TaskStore:     taskStore,
			LogStore:      &storemock.LogStoreMock{},
			StatusGetter:  statusGetter,
			Updater:       updater,
		})
		return handler, deviceStore, firewallStore, taskStore, statusGetter, updater
	}

	t.Run("page renders device, rules and tasks", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, taskStore, statusGetter, _ := newHandler()
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("ListRules", uint(1)).Return([]store.FirewallRule{
			{ID: 5, Direction: "inbound", Protocol: "tcp", PortFrom: 3389, PortTo: 3389, Action: "allow", Enabled: true},
		}, nil)
		taskStore.On("ListTasksForDevice", uint(1)).Return([]store.UpdateTask{
			{ID: 3, PackageName: "office-agent", PackageVersion: "2.0.0", Status: "completed"},
		}, nil)
		statusGetter.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
			Name: "lab-pc-1", Version: "2.0.0", OS: "NixOS 25.05", UptimeSeconds: 60,
		}, nil)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, deviceRequest(admin, true, "GET", "/devices/1", map[string]string{"id": "1"}, nil))

		assert.Equal(http.StatusOK, rr.Code)
		for _, expected := range []string{"lab-pc-1", "3389", "office-agent", "2.0.0"} {
			assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q", expected)
		}
		deviceStore.AssertExpectations(t)
	})

	t.Run("missing device is a 404", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, _, _, _ := newHandler()
		deviceStore.On("GetDevice", uint(9)).Return(nil, errors.New("not found"))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, deviceRequest(admin, true, "GET", "/devices/9", map[string]string{"id": "9"}, nil))

		assert.Equal(http.StatusNotFound, rr.Code)
		deviceStore.AssertExpectations(t)
	})

	t.Run("anonymous user redirects to login", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _, _, _ := newHandler()

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, deviceRequest(nil, false, "GET", "/devices/1", map[string]string{"id": "1"}, nil))

		assert.Equal(http.StatusFound, rr.Code)
		assert.Equal("/login", rr.Header().Get("Location"))
	})

	t.Run("add valid rule redirects", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, _, _, _ := newHandler()
		logStore := &storemock.LogStoreMock{}
		handler.logStore = logStore
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("AddRule", mock.MatchedBy(func(r *store.FirewallRule) bool {
			return r.DeviceID == 1 && r.Direction == "inbound" && r.Protocol == "tcp" && r.PortFrom == 22 && r.PortTo == 22 && r.Action == "allow" && r.Enabled
		})).Return(&store.FirewallRule{ID: 6}, nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		form := url.Values{"direction": {"inbound"}, "protocol": {"tcp"}, "port_from": {"22"}, "port_to": {"22"}, "action": {"allow"}}
		rr := httptest.NewRecorder()
		handler.AddRule(rr, deviceRequest(admin, true, "POST", "/devices/1/firewall", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusSeeOther, rr.Code)
		deviceStore.AssertExpectations(t)
		firewallStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("add invalid rule redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, _, _, _ := newHandler()
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)

		form := url.Values{"direction": {"sideways"}, "protocol": {"tcp"}, "port_from": {"22"}, "port_to": {"22"}, "action": {"allow"}}
		rr := httptest.NewRecorder()
		handler.AddRule(rr, deviceRequest(admin, true, "POST", "/devices/1/firewall", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "applied=invalid")
		deviceStore.AssertExpectations(t)
	})

	t.Run("toggle rule flips enabled", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, _, _, _ := newHandler()
		logStore := &storemock.LogStoreMock{}
		handler.logStore = logStore
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("SetRuleEnabled", uint(5), false).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		form := url.Values{"enabled": {"0"}}
		rr := httptest.NewRecorder()
		handler.ToggleRule(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1", "ruleID": "5"}, form))

		assert.Equal(http.StatusSeeOther, rr.Code)
		firewallStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("delete rule redirects", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, _, _, _ := newHandler()
		logStore := &storemock.LogStoreMock{}
		handler.logStore = logStore
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("DeleteRule", uint(5)).Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		rr := httptest.NewRecorder()
		handler.DeleteRule(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1", "ruleID": "5"}, nil))

		assert.Equal(http.StatusSeeOther, rr.Code)
		firewallStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("apply success and failure redirect with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, _, _, updater := newHandler()
		logStore := &storemock.LogStoreMock{}
		handler.logStore = logStore
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("ListRules", uint(1)).Return([]store.FirewallRule{
			{ID: 5, DeviceID: 1, Direction: "inbound", Protocol: "tcp", PortFrom: 3389, PortTo: 3389, Action: "allow", Enabled: true},
		}, nil)
		updater.On("ApplyFirewall", "http://localhost:4101", mock.Anything).Return(nil).Once()
		logStore.On("AddEntry", mock.Anything).Return(nil)

		rr := httptest.NewRecorder()
		handler.ApplyRules(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1"}, nil))
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "applied=ok")

		updater.On("ApplyFirewall", "http://localhost:4101", mock.Anything).Return(errors.New("down")).Once()
		rr = httptest.NewRecorder()
		handler.ApplyRules(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1"}, nil))
		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "applied=fail")

		updater.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("non-admin cannot add rules", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _, _, _ := newHandler()
		user := &store.User{ID: 2, Email: "user@example.com"}

		rr := httptest.NewRecorder()
		handler.AddRule(rr, deviceRequest(user, false, "POST", "/x", map[string]string{"id": "1"}, nil))

		assert.Equal(http.StatusFound, rr.Code)
	})

	t.Run("power reboot redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, _, _, updater := newHandler()
		logStore := &storemock.LogStoreMock{}
		handler.logStore = logStore
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		updater.On("SendPower", "http://localhost:4101", "reboot").Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		form := url.Values{"action": {"reboot"}}
		rr := httptest.NewRecorder()
		handler.PowerDevice(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "power=ok")
		updater.AssertExpectations(t)
		logStore.AssertExpectations(t)
	})

	t.Run("power bad action is rejected", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, _, _, _ := newHandler()
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)

		form := url.Values{"action": {"overclock"}}
		rr := httptest.NewRecorder()
		handler.PowerDevice(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusBadRequest, rr.Code)
		deviceStore.AssertExpectations(t)
	})

	t.Run("non-admin cannot power", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _, _, _ := newHandler()
		user := &store.User{ID: 2, Email: "user@example.com"}

		form := url.Values{"action": {"reboot"}}
		rr := httptest.NewRecorder()
		handler.PowerDevice(rr, deviceRequest(user, false, "POST", "/x", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusFound, rr.Code)
	})

	t.Run("console renders command output", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, taskStore, statusGetter, updater := newHandler()
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("ListRules", uint(1)).Return([]store.FirewallRule{}, nil)
		taskStore.On("ListTasksForDevice", uint(1)).Return([]store.UpdateTask{}, nil)
		statusGetter.On("GetStatus", "http://localhost:4101").Return(&agent.Status{
			Name: "lab-pc-1", Version: "2.0.0", OS: "NixOS 25.05", UptimeSeconds: 60,
		}, nil)
		updater.On("SendCommand", "http://localhost:4101", "whoami").Return("lab-pc-1", nil)

		form := url.Values{"command": {"whoami"}}
		rr := httptest.NewRecorder()
		handler.RunConsole(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusOK, rr.Code)
		assert.True(bytes.Contains(rr.Body.Bytes(), []byte("lab-pc-1")))
		updater.AssertExpectations(t)
	})

	t.Run("console shows error when device is down", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, firewallStore, taskStore, statusGetter, updater := newHandler()
		logStore := &storemock.LogStoreMock{}
		handler.logStore = logStore
		deviceStore.On("GetDevice", uint(1)).Return(device, nil)
		firewallStore.On("ListRules", uint(1)).Return([]store.FirewallRule{}, nil)
		taskStore.On("ListTasksForDevice", uint(1)).Return([]store.UpdateTask{}, nil)
		statusGetter.On("GetStatus", "http://localhost:4101").Return(nil, errors.New("down"))
		updater.On("SendCommand", "http://localhost:4101", "whoami").Return("", errors.New("down"))
		logStore.On("AddEntry", mock.Anything).Return(nil)

		form := url.Values{"command": {"whoami"}}
		rr := httptest.NewRecorder()
		handler.RunConsole(rr, deviceRequest(admin, true, "POST", "/x", map[string]string{"id": "1"}, form))

		assert.Equal(http.StatusOK, rr.Code)
		assert.True(bytes.Contains(rr.Body.Bytes(), []byte("did not respond")))
		updater.AssertExpectations(t)
	})
}
