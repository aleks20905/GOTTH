package handlers

import (
	"bytes"
	"errors"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type updateMock struct {
	mock.Mock
}

func (m *updateMock) StartUpdate(baseURL string, pkg string, version string) error {
	args := m.Called(baseURL, pkg, version)
	return args.Error(0)
}

func TestUpdatesPage(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	t.Run("page renders tasks with resolved targets", func(t *testing.T) {
		assert := assert.New(t)
		deviceStore := &storemock.DeviceStoreMock{}
		groupStore := &storemock.GroupStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		deviceID := uint(1)
		groupID := uint(2)
		taskStore.On("ListTasks").Return([]store.UpdateTask{
			{ID: 1, PackageName: "office-agent", PackageVersion: "2.0.0", DeviceID: &deviceID, Status: "running", CreatedBy: "boss@example.com"},
			{ID: 2, PackageName: "office-agent", PackageVersion: "2.0.0", GroupID: &groupID, Status: "pending", CreatedBy: "boss@example.com"},
		}, nil)
		deviceStore.On("ListDevices").Return([]store.Device{{ID: 1, Name: "lab-pc-1"}}, nil)
		groupStore.On("ListGroups").Return([]store.DeviceGroup{{ID: 2, Name: "Finance"}}, nil)

		handler := NewUpdatesHandler(UpdatesHandlerParams{
			DeviceStore: deviceStore,
			GroupStore:  groupStore,
			TaskStore:   taskStore,
			LogStore:    &storemock.LogStoreMock{},
			Updater:     &updateMock{},
		})

		req, _ := http.NewRequest("GET", "/admin/updates", nil)
		req = req.WithContext(adminTestContext(admin, true))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusOK, rr.Code)
		for _, expected := range []string{"Updates", "lab-pc-1", "Group: Finance", "Running", "Pending"} {
			assert.True(bytes.Contains(rr.Body.Bytes(), []byte(expected)), "body missing %q", expected)
		}
		taskStore.AssertExpectations(t)
	})

	t.Run("non-admin redirects home", func(t *testing.T) {
		assert := assert.New(t)
		handler := NewUpdatesHandler(UpdatesHandlerParams{})

		req, _ := http.NewRequest("GET", "/admin/updates", nil)
		req = req.WithContext(adminTestContext(&store.User{ID: 2, Email: "user@example.com"}, false))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(http.StatusFound, rr.Code)
	})
}

func TestCreateTask(t *testing.T) {
	admin := &store.User{ID: 1, Email: "boss@example.com"}

	newHandler := func() (*UpdatesHandler, *storemock.DeviceStoreMock, *storemock.GroupStoreMock, *storemock.TaskStoreMock, *storemock.LogStoreMock, *updateMock) {
		deviceStore := &storemock.DeviceStoreMock{}
		groupStore := &storemock.GroupStoreMock{}
		taskStore := &storemock.TaskStoreMock{}
		logStore := &storemock.LogStoreMock{}
		updater := &updateMock{}
		handler := NewUpdatesHandler(UpdatesHandlerParams{
			DeviceStore: deviceStore,
			GroupStore:  groupStore,
			TaskStore:   taskStore,
			LogStore:    logStore,
			Updater:     updater,
		})
		return handler, deviceStore, groupStore, taskStore, logStore, updater
	}

	post := func(form url.Values) *http.Request {
		req, _ := http.NewRequest("POST", "/admin/updates", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		return req.WithContext(adminTestContext(admin, true))
	}

	t.Run("immediate task dispatches at once", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, taskStore, logStore, updater := newHandler()
		deviceStore.On("ListDevices").Return([]store.Device{{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"}}, nil)
		taskStore.On("AddTask", mock.MatchedBy(func(task *store.UpdateTask) bool {
			return task.PackageName == "office-agent" && task.PackageVersion == "2.0.0" && task.Status == "pending" && task.DeviceID != nil && *task.DeviceID == 1
		})).Return(&store.UpdateTask{ID: 9, PackageName: "office-agent", PackageVersion: "2.0.0"}, nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)
		updater.On("StartUpdate", "http://localhost:4101", "office-agent", "2.0.0").Return(nil)
		taskStore.On("SetTaskStatus", uint(9), "running").Return(nil)

		rr := httptest.NewRecorder()
		handler.CreateTask(rr, post(url.Values{"package": {"office-agent"}, "version": {"2.0.0"}, "device_id": {"1"}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		taskStore.AssertExpectations(t)
		logStore.AssertExpectations(t)
		updater.AssertExpectations(t)
	})

	t.Run("future task stays pending", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, taskStore, logStore, updater := newHandler()
		taskStore.On("AddTask", mock.Anything).Return(&store.UpdateTask{ID: 10}, nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)

		rr := httptest.NewRecorder()
		handler.CreateTask(rr, post(url.Values{
			"package": {"office-agent"}, "version": {"2.0.0"}, "device_id": {"1"},
			"scheduled_at": {"2999-01-01T10:00"},
		}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		taskStore.AssertExpectations(t)
		updater.AssertExpectations(t)
	})

	t.Run("missing package redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _, _, _ := newHandler()

		rr := httptest.NewRecorder()
		handler.CreateTask(rr, post(url.Values{"version": {"2.0.0"}, "device_id": {"1"}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "error=empty")
	})

	t.Run("missing target redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _, _, _ := newHandler()

		rr := httptest.NewRecorder()
		handler.CreateTask(rr, post(url.Values{"package": {"office-agent"}, "version": {"2.0.0"}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "error=target")
	})

	t.Run("bad time redirects with flag", func(t *testing.T) {
		assert := assert.New(t)
		handler, _, _, _, _, _ := newHandler()

		rr := httptest.NewRecorder()
		handler.CreateTask(rr, post(url.Values{
			"package": {"office-agent"}, "version": {"2.0.0"}, "device_id": {"1"},
			"scheduled_at": {"yesterday"},
		}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		assert.Contains(rr.Header().Get("Location"), "error=time")
	})

	t.Run("unreachable device fails the task", func(t *testing.T) {
		assert := assert.New(t)
		handler, deviceStore, _, taskStore, logStore, updater := newHandler()
		deviceStore.On("ListDevices").Return([]store.Device{{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"}}, nil)
		taskStore.On("AddTask", mock.Anything).Return(&store.UpdateTask{ID: 11}, nil)
		taskStore.On("SetTaskStatus", uint(11), "failed").Return(nil)
		logStore.On("AddEntry", mock.Anything).Return(nil)
		updater.On("StartUpdate", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("down"))

		rr := httptest.NewRecorder()
		handler.CreateTask(rr, post(url.Values{"package": {"office-agent"}, "version": {"2.0.0"}, "device_id": {"1"}}))

		assert.Equal(http.StatusSeeOther, rr.Code)
		taskStore.AssertExpectations(t)
	})
}
