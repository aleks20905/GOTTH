package dbstore

import (
	"errors"
	"goth/internal/hash"
	"goth/internal/store"

	"gorm.io/gorm"
)

var (
	ErrUserExists   = errors.New("user with this email already exists")
	ErrUserNotFound = errors.New("user not found")
)

type UserStore struct {
	db           *gorm.DB
	passwordhash hash.PasswordHash
}

type NewUserStoreParams struct {
	DB           *gorm.DB
	PasswordHash hash.PasswordHash
}

func NewUserStore(params NewUserStoreParams) *UserStore {
	return &UserStore{
		db:           params.DB,
		passwordhash: params.PasswordHash,
	}
}

func (s *UserStore) CreateUser(email string, password string) error {
	// Check if email already exists
	var existing store.User
	err := s.db.Where("email = ?", email).First(&existing).Error

	if err == nil {
		// Found a user - email is taken
		return ErrUserExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other database error
		return err
	}

	// Email is available - create user
	hashedPassword, err := s.passwordhash.GenerateFromPassword(password)
	if err != nil {
		return err
	}

	return s.db.Create(&store.User{
		Email:    email,
		Password: hashedPassword,
	}).Error
}

func (s *UserStore) GetUser(email string) (*store.User, error) {
	var user store.User
	err := s.db.Where("email = ?", email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}
