package router

import (
	"net/http"

	"goth/internal/config"
	"goth/internal/handlers"
	"goth/internal/hash/passwordhash"
	"goth/internal/middleware"
	"goth/internal/store/dbstore"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type RouterDependencies struct {
	Config         config.Config
	UserStore      *dbstore.UserStore
	SessionStore   *dbstore.SessionStore
	PasswordHasher *passwordhash.PasswordHash
	ProductStore   *dbstore.ProductStore
	CartStore      *dbstore.CartStore
}

func SetupRouter(deps RouterDependencies) *chi.Mux {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir(deps.Config.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	authMiddleware := middleware.NewAuthMiddleware(deps.SessionStore, deps.Config.SessionCookieName)

	r.Group(func(r chi.Router) {
		r.Use(
			chiMiddleware.Logger,
			middleware.TextHTMLMiddleware,
			middleware.CSPMiddleware,
			authMiddleware.AddUserToContext,
		)

		r.NotFound(handlers.NewNotFoundHandler().ServeHTTP)

		r.Get("/", handlers.NewHomeHandler().ServeHTTP)

		r.Get("/about", handlers.NewAboutHandler().ServeHTTP)

		r.Get("/account", handlers.NewAccountHandler().ServeHTTP)

		// Product routes
		productListHandler := handlers.NewProductListHandler(deps.ProductStore)
		productHandler := handlers.NewProductHandler(deps.ProductStore)
		// Routes
		r.Get("/products", productListHandler.ServeHTTP)
		r.Get("/products/{id}", productHandler.ServeHTTP)

		cartHandler := handlers.NewCartHandler(handlers.NewCartHandlerParams{
			CartStore:    deps.CartStore,
			ProductStore: deps.ProductStore,
		})
		// Routes
		r.Get("/cart", cartHandler.ServeHTTP)
		r.Post("/cart/add/{id}", cartHandler.AddToCart)
		r.Post("/cart/remove/{id}", cartHandler.RemoveFromCart)
		r.Post("/cart/update/{id}", cartHandler.UpdateQuantity)
		r.Get("/cart/badge", cartHandler.GetBadge)

		r.Get("/quantity", cartHandler.UpdateQuantitySelector)

		r.Get("/register", handlers.NewGetRegisterHandler().ServeHTTP)

		r.Post("/register", handlers.NewPostRegisterHandler(handlers.PostRegisterHandlerParams{
			UserStore: deps.UserStore,
		}).ServeHTTP)

		r.Get("/login", handlers.NewGetLoginHandler().ServeHTTP)

		r.Post("/login", handlers.NewPostLoginHandler(handlers.PostLoginHandlerParams{
			UserStore:         deps.UserStore,
			SessionStore:      deps.SessionStore,
			PasswordHash:      deps.PasswordHasher,
			SessionCookieName: deps.Config.SessionCookieName,
		}).ServeHTTP)

		r.Post("/logout", handlers.NewPostLogoutHandler(handlers.PostLogoutHandlerParams{
			SessionCookieName: deps.Config.SessionCookieName,
		}).ServeHTTP)
	})

	return r
}
