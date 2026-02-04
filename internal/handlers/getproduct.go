package handlers

import (
	"goth/internal/store/dbstore"
	"goth/internal/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	ProductStore *dbstore.ProductStore
}

func NewProductHandler(productStore *dbstore.ProductStore) *ProductHandler {
	return &ProductHandler{
		ProductStore: productStore,
	}
}

func (h *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		// Show 404
		NewNotFoundHandler().ServeHTTP(w, r)
		return
	}

	product, err := h.ProductStore.GetProductByID(uint(id))
	if err != nil {
		// Product not found
		NewNotFoundHandler().ServeHTTP(w, r)
		return
	}

	c := templates.ProductDetail(*product)
	err = templates.Layout(c, product.Name).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
