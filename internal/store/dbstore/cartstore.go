package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type CartStore struct {
	db *gorm.DB
}

type NewCartStoreParams struct {
	DB *gorm.DB
}

func NewCartStore(params NewCartStoreParams) *CartStore {
	return &CartStore{
		db: params.DB,
	}
}

// GetOrCreateCart finds existing cart or creates new one
func (s *CartStore) GetOrCreateCart(userID *uint, sessionID string) (*store.Cart, error) {
	var cart store.Cart

	// Try to find existing cart
	query := s.db.Preload("Items.Product")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("session_id = ? AND user_id IS NULL", sessionID)
	}

	err := query.First(&cart).Error

	if err == gorm.ErrRecordNotFound {
		// Create new cart
		cart = store.Cart{
			UserID:    userID,
			SessionID: sessionID,
		}
		err = s.db.Create(&cart).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &cart, nil
}

// AddItem adds a product to cart or increases quantity
func (s *CartStore) AddItem(cartID uint, productID uint, quantity int) error {
	var item store.CartItem

	err := s.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error

	if err == gorm.ErrRecordNotFound {
		// Create new item
		item = store.CartItem{
			CartID:    cartID,
			ProductID: productID,
			Quantity:  quantity,
		}
		return s.db.Create(&item).Error
	} else if err != nil {
		return err
	}

	// Update quantity
	item.Quantity += quantity
	return s.db.Save(&item).Error
}

// UpdateItemQuantity sets the quantity directly
func (s *CartStore) UpdateItemQuantity(cartID uint, productID uint, quantity int) error {
	if quantity <= 0 {
		return s.RemoveItem(cartID, productID)
	}

	return s.db.Model(&store.CartItem{}).
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		Update("quantity", quantity).Error
}

// RemoveItem removes a product from cart
func (s *CartStore) RemoveItem(cartID uint, productID uint) error {
	return s.db.Where("cart_id = ? AND product_id = ?", cartID, productID).
		Delete(&store.CartItem{}).Error
}

// GetCartWithItems loads cart with all items and products
func (s *CartStore) GetCartWithItems(cartID uint) (*store.Cart, error) {
	var cart store.Cart
	err := s.db.Preload("Items.Product").First(&cart, cartID).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// ClearCart removes all items from cart
func (s *CartStore) ClearCart(cartID uint) error {
	return s.db.Where("cart_id = ?", cartID).Delete(&store.CartItem{}).Error
}

// GetCartTotals calculates totals (helper method)
func (s *CartStore) GetCartTotals(cart *store.Cart) (totalItems int, totalPrice float64) {
	for _, item := range cart.Items {
		totalItems += item.Quantity
		totalPrice += item.Product.Price * float64(item.Quantity)
	}
	return
}

// MergeGuestCart merges guest cart into user cart after login
func (s *CartStore) MergeGuestCart(userID uint, sessionID string) error {
	var guestCart store.Cart

	// Find guest cart
	err := s.db.Preload("Items").
		Where("session_id = ? AND user_id IS NULL", sessionID).
		First(&guestCart).Error

	if err == gorm.ErrRecordNotFound {
		return nil // No guest cart to merge
	} else if err != nil {
		return err
	}

	// Get or create user cart
	userCart, err := s.GetOrCreateCart(&userID, "")
	if err != nil {
		return err
	}

	// Merge items
	for _, item := range guestCart.Items {
		s.AddItem(userCart.ID, item.ProductID, item.Quantity)
	}

	// Delete guest cart
	s.db.Delete(&guestCart)

	return nil
}
