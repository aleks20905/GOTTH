package scheduler

import (
	"errors"
	"goth/internal/store"
	storemock "goth/internal/store/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type updaterMock struct {
	mock.Mock
}

func (m *updaterMock) StartUpdate(baseURL string, pkg string, version string) error {
	args := m.Called(baseURL, pkg, version)
	return args.Error(0)
}

func TestTickDispatchesDueTasks(t *testing.T) {
	now := time.Now()
	deviceStore := &storemock.DeviceStoreMock{}
	taskStore := &storemock.TaskStoreMock{}
	logStore := &storemock.LogStoreMock{}
	updater := &updaterMock{}

	deviceID := uint(1)
	taskStore.On("ListTasksByStatus", "pending").Return([]store.UpdateTask{
		{ID: 1, PackageName: "office-agent", PackageVersion: "2.0.0", DeviceID: &deviceID, Status: "pending", ScheduledAt: now.Add(-time.Minute), CreatedBy: "boss@example.com"},
		{ID: 2, PackageName: "office-agent", PackageVersion: "2.0.0", DeviceID: &deviceID, Status: "pending", ScheduledAt: now.Add(time.Hour), CreatedBy: "boss@example.com"},
	}, nil)
	taskStore.On("ListTasksByStatus", "running").Return([]store.UpdateTask{}, nil)
	deviceStore.On("ListDevices").Return([]store.Device{
		{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"},
	}, nil)
	updater.On("StartUpdate", "http://localhost:4101", "office-agent", "2.0.0").Return(nil)
	taskStore.On("SetTaskStatus", uint(1), "running").Return(nil)
	logStore.On("AddEntry", mock.Anything).Return(nil)

	Tick(deviceStore, taskStore, logStore, updater, now)

	deviceStore.AssertExpectations(t)
	taskStore.AssertExpectations(t)
	logStore.AssertExpectations(t)
	updater.AssertExpectations(t)
}

func TestTickFailsUnreachableTasks(t *testing.T) {
	now := time.Now()
	deviceStore := &storemock.DeviceStoreMock{}
	taskStore := &storemock.TaskStoreMock{}
	logStore := &storemock.LogStoreMock{}
	updater := &updaterMock{}

	deviceID := uint(1)
	taskStore.On("ListTasksByStatus", "pending").Return([]store.UpdateTask{
		{ID: 1, PackageName: "office-agent", PackageVersion: "2.0.0", DeviceID: &deviceID, Status: "pending", ScheduledAt: now.Add(-time.Minute), CreatedBy: "boss@example.com"},
	}, nil)
	taskStore.On("ListTasksByStatus", "running").Return([]store.UpdateTask{}, nil)
	deviceStore.On("ListDevices").Return([]store.Device{
		{ID: 1, Name: "lab-pc-1", URL: "http://localhost:4101"},
	}, nil)
	updater.On("StartUpdate", "http://localhost:4101", "office-agent", "2.0.0").Return(errors.New("down"))
	taskStore.On("SetTaskStatus", uint(1), "failed").Return(nil)
	logStore.On("AddEntry", mock.Anything).Return(nil)

	Tick(deviceStore, taskStore, logStore, updater, now)

	taskStore.AssertExpectations(t)
	logStore.AssertExpectations(t)
}

func TestTickFailsStaleRunningTasks(t *testing.T) {
	now := time.Now()
	deviceStore := &storemock.DeviceStoreMock{}
	taskStore := &storemock.TaskStoreMock{}
	logStore := &storemock.LogStoreMock{}
	updater := &updaterMock{}

	taskStore.On("ListTasksByStatus", "pending").Return([]store.UpdateTask{}, nil)
	taskStore.On("ListTasksByStatus", "running").Return([]store.UpdateTask{
		{ID: 1, PackageName: "office-agent", PackageVersion: "2.0.0", Status: "running", CreatedBy: "boss@example.com", UpdatedAt: now.Add(-11 * time.Minute)},
		{ID: 2, PackageName: "office-agent", PackageVersion: "2.0.0", Status: "running", CreatedBy: "boss@example.com", UpdatedAt: now.Add(-time.Minute)},
	}, nil)
	taskStore.On("SetTaskStatus", uint(1), "failed").Return(nil)
	logStore.On("AddEntry", mock.Anything).Return(nil)

	Tick(deviceStore, taskStore, logStore, updater, now)

	taskStore.AssertExpectations(t)
	logStore.AssertExpectations(t)
	updater.AssertExpectations(t)
}
