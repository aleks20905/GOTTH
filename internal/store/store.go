package store

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `json:"-"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `json:"session_id"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Device struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	Name      string       `json:"name"`
	URL       string       `gorm:"uniqueIndex" json:"url"`
	Version   string       `json:"version"`
	OS        string       `json:"os"`
	Verified  bool         `json:"verified"`
	LastSeen  time.Time    `json:"last_seen"`
	GroupID   *uint        `json:"group_id"`
	Group     *DeviceGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

type DeviceGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type FirewallRule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DeviceID      uint      `gorm:"index" json:"device_id"`
	Direction     string    `json:"direction"`
	Protocol      string    `json:"protocol"`
	PortFrom      int       `json:"port_from"`
	PortTo        int       `json:"port_to"`
	RemoteAddress string    `json:"remote_address"`
	Action        string    `json:"action"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
}

type UpdateTask struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PackageName    string    `json:"package_name"`
	PackageVersion string    `json:"package_version"`
	DeviceID       *uint     `json:"device_id"`
	GroupID        *uint     `json:"group_id"`
	Status         string    `json:"status"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LogEntry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Details   string    `json:"details"`
	UserEmail string    `json:"user_email"`
}

type UserStore interface {
	CreateUser(email string, password string) error
	GetUser(email string) (*User, error)
	ListUsers() ([]User, error)
	UpdateEmail(id uint, email string) error
	UpdatePassword(id uint, hash string) error
	SetAdmin(id uint, admin bool) error
}
type SessionStore interface {
	CreateSession(session *Session) (*Session, error)
	GetUserFromSession(sessionID string, userID string) (*User, error)
}
type DeviceStore interface {
	ListDevices() ([]Device, error)
	GetDevice(id uint) (*Device, error)
	AddDevice(device *Device) (*Device, error)
	UpsertDevice(device *Device) (*Device, error)
	AssignGroup(deviceID uint, groupID *uint) error
	DeleteDevice(id uint) error
}
type GroupStore interface {
	ListGroups() ([]DeviceGroup, error)
	AddGroup(group *DeviceGroup) (*DeviceGroup, error)
	DeleteGroup(id uint) error
}
type FirewallStore interface {
	ListRules(deviceID uint) ([]FirewallRule, error)
	AddRule(rule *FirewallRule) (*FirewallRule, error)
	SetRuleEnabled(id uint, enabled bool) error
	DeleteRule(id uint) error
}
type TaskStore interface {
	ListTasks() ([]UpdateTask, error)
	ListTasksByStatus(status string) ([]UpdateTask, error)
	ListTasksForDevice(deviceID uint) ([]UpdateTask, error)
	AddTask(task *UpdateTask) (*UpdateTask, error)
	SetTaskStatus(id uint, status string) error
}
type LogStore interface {
	ListEntries(limit int) ([]LogEntry, error)
	AddEntry(entry *LogEntry) error
}
