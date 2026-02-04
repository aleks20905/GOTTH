package handlers

import (
	"goth/internal/store/dbstore"
	"goth/internal/templates"
	"net/http"
)

type ProductListHandler struct {
	ProductStore *dbstore.ProductStore
}

func NewProductListHandler(productStore *dbstore.ProductStore) *ProductListHandler {
	return &ProductListHandler{
		ProductStore: productStore,
	}
}

func (h *ProductListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductStore.GetAllProducts()
	if err != nil {
		http.Error(w, "Failed to load products", http.StatusInternalServerError)
		return
	}

	c := templates.Products(products)
	err = templates.Layout(c, "Products").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
