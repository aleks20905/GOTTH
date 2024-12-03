package router

import (
	"goth/internal/config"
	"goth/internal/handlers"
	"goth/internal/hash/passwordhash"
	"goth/internal/middleware"
	"goth/internal/store/dbstore"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(cfg config.Config, userStore *dbstore.UserStore, sessionStore *dbstore.SessionStore, passwordHasher *passwordhash.PasswordHash, scheduleStore *dbstore.ScheduleStore) *chi.Mux {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	authMiddleware := middleware.NewAuthMiddleware(sessionStore, cfg.SessionCookieName)

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

		r.Get("/weekly", handlers.NewWeeklyHandler(handlers.GetWeeklyHandlerParams{
			ScheduleStore: scheduleStore,
		}).ServeHTTP)

		r.Get("/register", handlers.NewGetRegisterHandler().ServeHTTP)

		r.Post("/register", handlers.NewPostRegisterHandler(handlers.PostRegisterHandlerParams{
			UserStore: userStore,
		}).ServeHTTP)

		r.Get("/login", handlers.NewGetLoginHandler().ServeHTTP)

		r.Post("/login", handlers.NewPostLoginHandler(handlers.PostLoginHandlerParams{
			UserStore:         userStore,
			SessionStore:      sessionStore,
			PasswordHash:      passwordHasher,
			SessionCookieName: cfg.SessionCookieName,
		}).ServeHTTP)

		r.Post("/logout", handlers.NewPostLogoutHandler(handlers.PostLogoutHandlerParams{
			SessionCookieName: cfg.SessionCookieName,
		}).ServeHTTP)
	})

	return r
}
