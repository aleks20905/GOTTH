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
type SessionStore interface {
	CreateSession(session *Session) (*Session, error)
	GetUserFromSession(sessionID string, userID string) (*User, error)
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
