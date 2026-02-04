package store

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"` // 👈 Unique index
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

type Cart struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    *uint      `json:"user_id"` // nil for guest carts
	User      *User      `gorm:"foreignKey:UserID" json:"user"`
	SessionID string     `json:"session_id"` // for guest users
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CartItem represents an item in the cart
type CartItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	CartID    uint    `json:"cart_id"`
	ProductID uint    `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int     `json:"quantity"`
}

// Product - your actual products
type Product struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Weight      string  `json:"weight"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	GetSubjectQuestions(subject string) (*SubjectQuestions, error)
	GetAllSubjects() ([]string, error)
	GetCorrectAnswers(subject string, questionID int) ([]string, error)
}
type CartStore interface {
	GetOrCreateCart(userID *uint, sessionID string) (*Cart, error)
	AddItem(cartID uint, productID uint, quantity int) error
	UpdateItemQuantity(cartID uint, productID uint, quantity int) error
	RemoveItem(cartID uint, productID uint) error
	GetCartWithItems(cartID uint) (*Cart, error)
	ClearCart(cartID uint) error
}

type ProductStore interface {
	GetAllProducts() ([]Product, error)
	GetProductByID(id uint) (*Product, error)
	GetProductsByCategory(category string) ([]Product, error)
}
