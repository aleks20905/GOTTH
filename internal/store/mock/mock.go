package mock

import (
	"goth/internal/store"

	"github.com/stretchr/testify/mock"
)

type UserStoreMock struct {
	mock.Mock
}

func (m *UserStoreMock) CreateUser(email string, password string) error {
	args := m.Called(email, password)

	return args.Error(0)
}

func (m *UserStoreMock) GetUser(email string) (*store.User, error) {
	args := m.Called(email)
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *UserStoreMock) ListUsers() ([]store.User, error) {
	args := m.Called()
	return args.Get(0).([]store.User), args.Error(1)
}

func (m *UserStoreMock) UpdateEmail(id uint, email string) error {
	args := m.Called(id, email)
	return args.Error(0)
}

func (m *UserStoreMock) UpdatePassword(id uint, hash string) error {
	args := m.Called(id, hash)
	return args.Error(0)
}

func (m *UserStoreMock) SetAdmin(id uint, admin bool) error {
	args := m.Called(id, admin)
	return args.Error(0)
}

type SessionStoreMock struct {
	mock.Mock
}

func (m *SessionStoreMock) CreateSession(session *store.Session) (*store.Session, error) {
	args := m.Called(session)
	return args.Get(0).(*store.Session), args.Error(1)
}

func (m *SessionStoreMock) GetUserFromSession(sessionID string, userID string) (*store.User, error) {
	args := m.Called(sessionID, userID)
	return args.Get(0).(*store.User), args.Error(1)
}

type DeviceStoreMock struct {
	mock.Mock
}

func (m *DeviceStoreMock) ListDevices() ([]store.Device, error) {
	args := m.Called()
	devices, _ := args.Get(0).([]store.Device)
	return devices, args.Error(1)
}

func (m *DeviceStoreMock) GetDevice(id uint) (*store.Device, error) {
	args := m.Called(id)
	device, _ := args.Get(0).(*store.Device)
	return device, args.Error(1)
}

func (m *DeviceStoreMock) AssignGroup(deviceID uint, groupID *uint) error {
	args := m.Called(deviceID, groupID)
	return args.Error(0)
}

func (m *DeviceStoreMock) AddDevice(device *store.Device) (*store.Device, error) {
	args := m.Called(device)
	return args.Get(0).(*store.Device), args.Error(1)
}

func (m *DeviceStoreMock) UpsertDevice(device *store.Device) (*store.Device, error) {
	args := m.Called(device)
	return args.Get(0).(*store.Device), args.Error(1)
}

func (m *DeviceStoreMock) DeleteDevice(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type GroupStoreMock struct {
	mock.Mock
}

func (m *GroupStoreMock) ListGroups() ([]store.DeviceGroup, error) {
	args := m.Called()
	return args.Get(0).([]store.DeviceGroup), args.Error(1)
}

func (m *GroupStoreMock) AddGroup(group *store.DeviceGroup) (*store.DeviceGroup, error) {
	args := m.Called(group)
	result, _ := args.Get(0).(*store.DeviceGroup)
	return result, args.Error(1)
}

func (m *GroupStoreMock) DeleteGroup(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type FirewallStoreMock struct {
	mock.Mock
}

func (m *FirewallStoreMock) ListRules(deviceID uint) ([]store.FirewallRule, error) {
	args := m.Called(deviceID)
	return args.Get(0).([]store.FirewallRule), args.Error(1)
}

func (m *FirewallStoreMock) AddRule(rule *store.FirewallRule) (*store.FirewallRule, error) {
	args := m.Called(rule)
	return args.Get(0).(*store.FirewallRule), args.Error(1)
}

func (m *FirewallStoreMock) SetRuleEnabled(id uint, enabled bool) error {
	args := m.Called(id, enabled)
	return args.Error(0)
}

func (m *FirewallStoreMock) DeleteRule(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type TaskStoreMock struct {
	mock.Mock
}

func (m *TaskStoreMock) ListTasks() ([]store.UpdateTask, error) {
	args := m.Called()
	return args.Get(0).([]store.UpdateTask), args.Error(1)
}

func (m *TaskStoreMock) ListTasksByStatus(status string) ([]store.UpdateTask, error) {
	args := m.Called(status)
	return args.Get(0).([]store.UpdateTask), args.Error(1)
}

func (m *TaskStoreMock) ListTasksForDevice(deviceID uint) ([]store.UpdateTask, error) {
	args := m.Called(deviceID)
	return args.Get(0).([]store.UpdateTask), args.Error(1)
}

func (m *TaskStoreMock) AddTask(task *store.UpdateTask) (*store.UpdateTask, error) {
	args := m.Called(task)
	return args.Get(0).(*store.UpdateTask), args.Error(1)
}

func (m *TaskStoreMock) SetTaskStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

type LogStoreMock struct {
	mock.Mock
}

func (m *LogStoreMock) ListEntries(limit int) ([]store.LogEntry, error) {
	args := m.Called(limit)
	return args.Get(0).([]store.LogEntry), args.Error(1)
}

func (m *LogStoreMock) AddEntry(entry *store.LogEntry) error {
	args := m.Called(entry)
	return args.Error(0)
}
