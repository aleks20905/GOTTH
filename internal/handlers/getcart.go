package handlers

import (
	"goth/internal/middleware"
	"goth/internal/store"
	"goth/internal/store/dbstore"
	"goth/internal/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CartHandler struct {
	CartStore    *dbstore.CartStore
	ProductStore *dbstore.ProductStore
}

type NewCartHandlerParams struct {
	CartStore    *dbstore.CartStore
	ProductStore *dbstore.ProductStore
}

func NewCartHandler(params NewCartHandlerParams) *CartHandler {
	return &CartHandler{
		CartStore:    params.CartStore,
		ProductStore: params.ProductStore,
	}
}

// GET /cart
func (h *CartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cart := h.getCart(w, r)
	totalItems, totalPrice := h.CartStore.GetCartTotals(cart)

	c := templates.CartPage(cart, totalItems, totalPrice)
	err := templates.Layout(c, "Shopping Cart").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// POST /cart/add/{id}
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Verify product exists
	_, err = h.ProductStore.GetProductByID(uint(productID))
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	cart := h.getCart(w, r)

	err = h.CartStore.AddItem(cart.ID, uint(productID), 1)
	if err != nil {
		http.Error(w, "Failed to add item", http.StatusInternalServerError)
		return
	}

	// Reload cart for updated totals
	cart, _ = h.CartStore.GetCartWithItems(cart.ID)
	totalItems, _ := h.CartStore.GetCartTotals(cart)

	templates.CartBadge(totalItems).Render(r.Context(), w)
}

// POST /cart/remove/{id}
func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	cart := h.getCart(w, r)
	h.CartStore.RemoveItem(cart.ID, uint(productID))

	// Reload and render
	cart, _ = h.CartStore.GetCartWithItems(cart.ID)
	totalItems, totalPrice := h.CartStore.GetCartTotals(cart)

	templates.CartContent(cart, totalItems, totalPrice).Render(r.Context(), w)
}

// POST /cart/update/{id}
func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	quantity, _ := strconv.Atoi(r.FormValue("quantity"))

	cart := h.getCart(w, r)
	h.CartStore.UpdateItemQuantity(cart.ID, uint(productID), quantity)

	// Reload and render
	cart, _ = h.CartStore.GetCartWithItems(cart.ID)
	totalItems, totalPrice := h.CartStore.GetCartTotals(cart)

	templates.CartContent(cart, totalItems, totalPrice).Render(r.Context(), w)
}

// Helper: Get or create cart from session/user
func (h *CartHandler) getCart(w http.ResponseWriter, r *http.Request) *store.Cart {
	// Check if logged in user
	user, ok := r.Context().Value(middleware.UserKey).(*store.User)

	var userID *uint
	var sessionID string

	if ok && user != nil {
		userID = &user.ID
	} else {
		sessionID = h.getOrCreateSessionID(w, r)
	}

	cart, err := h.CartStore.GetOrCreateCart(userID, sessionID)
	if err != nil {
		// Return empty cart on error
		return &store.Cart{Items: []store.CartItem{}}
	}

	return cart
}

// Helper: Manage cart session cookie
func (h *CartHandler) getOrCreateSessionID(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("cart_session")
	if err == nil {
		return cookie.Value
	}

	// Create new session ID
	sessionID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     "cart_session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return sessionID
}
