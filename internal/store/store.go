package store

import "time"

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Session struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	User      User   `gorm:"foreignKey:UserID" json:"user"`
}
type Schedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Course    uint      `json:"course"`
	Spec      string    `json:"spec"`
	GroupName uint      `json:"group_name"`
	Title     string    `json:"title"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Room      string    `json:"room"`
	Teacher   string    `json:"teacher"`
	Type      string    `json:"type"`
	GroupS    string    `json:"groups"`
	Des       string    `json:"des"`
}
type DayScheduels struct {
	Day     string
	Shedule []Schedule
}
type UserStore interface {
	CreateUser(email string, password string) error
	GetUser(email string) (*User, error)
}
type ScheduleStore interface {
	GetSchedule(course uint, spec string, group_name string) (*[]Schedule, error)
	GetCourses() (*[]Schedule, error)
	GetSpec() (*[]Schedule, error)
	GetGroupName() (*[]Schedule, error)
}
type SessionStore interface {
	CreateSession(session *Session) (*Session, error)
	GetUserFromSession(sessionID string, userID string) (*User, error)
}
