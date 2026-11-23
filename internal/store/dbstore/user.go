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
	var existing store.User
	err := s.db.Where("email = ?", email).First(&existing).Error

	if err == nil {
		return ErrUserExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

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

func (s *UserStore) ListUsers() ([]store.User, error) {
	var users []store.User
	err := s.db.Order("id").Find(&users).Error
	return users, err
}

func (s *UserStore) UpdateEmail(id uint, email string) error {
	var existing store.User
	err := s.db.Where("email = ?", email).First(&existing).Error

	if err == nil {
		return ErrUserExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.db.Model(&store.User{}).Where("id = ?", id).Update("email", email).Error
}

func (s *UserStore) UpdatePassword(id uint, hash string) error {
	return s.db.Model(&store.User{}).Where("id = ?", id).Update("password", hash).Error
}

func (s *UserStore) SetAdmin(id uint, admin bool) error {
	return s.db.Model(&store.User{}).Where("id = ?", id).Update("is_admin", admin).Error
}
