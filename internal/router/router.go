package router

import (
	"log"
	"net/http"

	"goth/internal/agent"
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
	DeviceStore    *dbstore.DeviceStore
	GroupStore     *dbstore.GroupStore
	FirewallStore  *dbstore.FirewallStore
	TaskStore      *dbstore.TaskStore
	LogStore       *dbstore.LogStore
}

func SetupRouter(deps RouterDependencies) *chi.Mux {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir(deps.Config.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	sessionSecret := deps.Config.SessionSecret
	if sessionSecret == "" {
		// shortcut: sessions die on every restart, set SESSION_SECRET in the flake for prod
		sessionSecret = middleware.GenerateRandomString(32)
		log.Println("SESSION_SECRET not set, using ephemeral key")
	}

	authMiddleware := middleware.NewAuthMiddleware(deps.SessionStore, deps.Config.SessionCookieName, sessionSecret, deps.Config.AdminEmail)

	if deps.Config.AgentToken == "" {
		log.Println("AGENT_TOKEN not set, agent auto-registration disabled")
	}

	// Machine-to-machine pairing endpoint, outside the login group on purpose.
	r.Post("/api/agents/register", handlers.NewRegisterHandler(handlers.RegisterHandlerParams{
		DeviceStore: deps.DeviceStore,
		TaskStore:   deps.TaskStore,
		LogStore:    deps.LogStore,
		AgentToken:  deps.Config.AgentToken,
	}).ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(
			chiMiddleware.Logger,
			middleware.TextHTMLMiddleware,
			middleware.CSPMiddleware,
			authMiddleware.AddUserToContext,
		)

		r.NotFound(handlers.NewNotFoundHandler().ServeHTTP)

		agentClient := agent.NewClient(deps.Config.AgentToken)

		r.Get("/", handlers.NewHomeHandler(handlers.HomeDataParams{
			DeviceStore:  deps.DeviceStore,
			TaskStore:    deps.TaskStore,
			LogStore:     deps.LogStore,
			StatusGetter: agentClient,
		}).ServeHTTP)

		r.Get("/account", handlers.NewAccountHandler().ServeHTTP)

		accountHandler := handlers.NewPostAccountHandler(handlers.PostAccountHandlerParams{
			UserStore:    deps.UserStore,
			PasswordHash: deps.PasswordHasher,
			LogStore:     deps.LogStore,
		})

		r.Post("/account/email", accountHandler.ChangeEmail)
		r.Post("/account/password", accountHandler.ChangePassword)

		r.Get("/computers", handlers.NewComputersHandler(handlers.ComputersHandlerParams{
			DeviceStore:  deps.DeviceStore,
			GroupStore:   deps.GroupStore,
			StatusGetter: agentClient,
		}).ServeHTTP)

		deviceHandler := handlers.NewDeviceHandler(handlers.DeviceHandlerParams{
			DeviceStore:   deps.DeviceStore,
			FirewallStore: deps.FirewallStore,
			TaskStore:     deps.TaskStore,
			LogStore:      deps.LogStore,
			StatusGetter:  agentClient,
			Updater:       agentClient,
		})

		r.Get("/devices/{id}", deviceHandler.ServeHTTP)
		r.Post("/devices/{id}/firewall", deviceHandler.AddRule)
		r.Post("/devices/{id}/firewall/{ruleID}/toggle", deviceHandler.ToggleRule)
		r.Post("/devices/{id}/firewall/{ruleID}/delete", deviceHandler.DeleteRule)
		r.Post("/devices/{id}/firewall/apply", deviceHandler.ApplyRules)
		r.Post("/devices/{id}/power", deviceHandler.PowerDevice)
		r.Post("/devices/{id}/console", deviceHandler.RunConsole)

		r.Get("/logs", handlers.NewLogsHandler(handlers.LogsHandlerParams{
			LogStore: deps.LogStore,
		}).ServeHTTP)

		adminHandler := handlers.NewAdminHandler(handlers.AdminHandlerParams{
			DeviceStore:  deps.DeviceStore,
			StatusGetter: agentClient,
			LogStore:     deps.LogStore,
		})

		r.Get("/admin", adminHandler.ServeHTTP)
		r.Post("/admin/devices", adminHandler.AddDevice)
		r.Post("/admin/devices/delete/{id}", adminHandler.DeleteDevice)

		usersHandler := handlers.NewUsersHandler(handlers.UsersHandlerParams{
			UserStore: deps.UserStore,
			LogStore:  deps.LogStore,
		})

		r.Get("/admin/users", usersHandler.ServeHTTP)
		r.Post("/admin/users/{id}/grant", usersHandler.GrantAdmin)
		r.Post("/admin/users/{id}/revoke", usersHandler.RevokeAdmin)

		groupsHandler := handlers.NewGroupsHandler(handlers.GroupsHandlerParams{
			DeviceStore: deps.DeviceStore,
			GroupStore:  deps.GroupStore,
			LogStore:    deps.LogStore,
		})

		r.Get("/admin/groups", groupsHandler.ServeHTTP)
		r.Post("/admin/groups", groupsHandler.AddGroup)
		r.Post("/admin/groups/{id}/delete", groupsHandler.DeleteGroup)
		r.Post("/admin/groups/{id}/assign", groupsHandler.AssignDevice)
		r.Post("/admin/groups/{id}/unassign/{deviceID}", groupsHandler.UnassignDevice)

		updatesHandler := handlers.NewUpdatesHandler(handlers.UpdatesHandlerParams{
			DeviceStore: deps.DeviceStore,
			GroupStore:  deps.GroupStore,
			TaskStore:   deps.TaskStore,
			LogStore:    deps.LogStore,
			Updater:     agentClient,
		})

		r.Get("/admin/updates", updatesHandler.ServeHTTP)
		r.Post("/admin/updates", updatesHandler.CreateTask)

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
			SessionSecret:     sessionSecret,
		}).ServeHTTP)

		r.Post("/logout", handlers.NewPostLogoutHandler(handlers.PostLogoutHandlerParams{
			SessionCookieName: deps.Config.SessionCookieName,
		}).ServeHTTP)
	})

	return r
}
