package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type ProductStore struct {
	db *gorm.DB
}

type NewProductStoreParams struct {
	DB *gorm.DB
}

func NewProductStore(params NewProductStoreParams) *ProductStore {
	return &ProductStore{
		db: params.DB,
	}
}

func (s *ProductStore) GetAllProducts() ([]store.Product, error) {
	var products []store.Product
	err := s.db.Find(&products).Error
	return products, err
}

func (s *ProductStore) GetProductByID(id uint) (*store.Product, error) {
	var product store.Product
	err := s.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) GetProductsByCategory(category string) ([]store.Product, error) {
	var products []store.Product
	err := s.db.Where("category = ?", category).Find(&products).Error
	return products, err
}
