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

// not hooked to the db just for misc stuff
type DayScheduels struct {
	Day     string
	Shedule []Schedule
}

// Question represents a multiple-choice question structure
type Question struct {
	ID       int      `json:"id"` // Auto-generated ID
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Answer   []string `json:"answer"` // Multiple correct answers
}

// OpenQuestion represents an open-ended question structure
type OpenQuestion struct {
	ID       int    `json:"id"` // Auto-generated ID
	Question string `json:"question"`
	Answer   string `json:"answer"` // kinda not used at all why does it even exist no IDK ...
}

// SubjectQuestions stores both multiple-choice and open-ended questions for a subject
type SubjectQuestions struct {
	MultipleChoice []Question
	OpenEnded      []OpenQuestion
}

type UserStore interface {
	CreateUser(email string, password string) error
	GetUser(email string) (*User, error)
}
type ScheduleStore interface {
	GetSchedules(course uint, spec string, group_name string) (*[]Schedule, error)
	GetCourses() (*[]Schedule, error)
	GetSpecs() (*[]Schedule, error)
	GetGroupNames() (*[]Schedule, error)
	GetAllscheduleUrls() (*[]Schedule, error)
}
type SessionStore interface {
	CreateSession(session *Session) (*Session, error)
	GetUserFromSession(sessionID string, userID string) (*User, error)
}
type QuestionStorer interface {
	GetSubjectQuestions(subject string) (*SubjectQuestions, bool)
	GetAllSubjects() []string
}
