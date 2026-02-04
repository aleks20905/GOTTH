package handlers

import (
	"errors"
	"goth/internal/store"
	"goth/internal/store/dbstore"
	"goth/internal/templates"
	"net/http"
)

type PostRegisterHandler struct {
	userStore store.UserStore
}

type PostRegisterHandlerParams struct {
	UserStore store.UserStore
}

func NewPostRegisterHandler(params PostRegisterHandlerParams) *PostRegisterHandler {
	return &PostRegisterHandler{
		userStore: params.UserStore,
	}
}

func (h *PostRegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Basic validation
	if email == "" || password == "" {
		c := templates.RegisterForm("Email and password are required", email)
		templates.Layout(c, "Register").Render(r.Context(), w)
		return
	}

	if len(password) < 8 {
		c := templates.RegisterForm("Password must be at least 8 characters", email)
		templates.Layout(c, "Register").Render(r.Context(), w)
		return
	}

	err := h.userStore.CreateUser(email, password)

	if errors.Is(err, dbstore.ErrUserExists) {
		c := templates.RegisterForm("An account with this email already exists", email)
		templates.Layout(c, "Register").Render(r.Context(), w)
		return
	}

	if err != nil {
		c := templates.RegisterForm("Something went wrong. Please try again.", email)
		templates.Layout(c, "Register").Render(r.Context(), w)
		return
	}

	// Success - show success page
	c := templates.RegisterSuccess()
	templates.Layout(c, "Success").Render(r.Context(), w)
}
