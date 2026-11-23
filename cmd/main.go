package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goth/internal/agent"
	"goth/internal/config"
	"goth/internal/hash/passwordhash"
	"goth/internal/router"
	"goth/internal/scheduler"
	database "goth/internal/store/db"
	"goth/internal/store/dbstore"
)

/*
* Set to production at build time
* used to determine what assets to load
 */
var Environment string

func init() {
	env := os.Getenv("ENV")
	if env == "" {
		Environment = "development"
		os.Setenv("ENV", Environment)
	} else {
		Environment = env
		os.Setenv("ENV", Environment)
	}
}
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.MustLoadConfig()

	db := database.MustOpen(cfg.DatabaseName, cfg.DatabaseURL)
	passwordHasher := passwordhash.NewHPasswordHash()

	userStore := dbstore.NewUserStore(dbstore.NewUserStoreParams{DB: db, PasswordHash: passwordHasher})

	sessionStore := dbstore.NewSessionStore(dbstore.NewSessionStoreParams{DB: db})

	deviceStore := dbstore.NewDeviceStore(dbstore.NewDeviceStoreParams{DB: db})
	groupStore := dbstore.NewGroupStore(dbstore.NewGroupStoreParams{DB: db})
	firewallStore := dbstore.NewFirewallStore(dbstore.NewFirewallStoreParams{DB: db})
	taskStore := dbstore.NewTaskStore(dbstore.NewTaskStoreParams{DB: db})
	logStore := dbstore.NewLogStore(dbstore.NewLogStoreParams{DB: db})

	deps := router.RouterDependencies{
		Config:         *cfg,
		UserStore:      userStore,
		SessionStore:   sessionStore,
		PasswordHasher: passwordHasher,
		DeviceStore:    deviceStore,
		GroupStore:     groupStore,
		FirewallStore:  firewallStore,
		TaskStore:      taskStore,
		LogStore:       logStore,
	}
	r := router.SetupRouter(deps)

	go runScheduler(deviceStore, taskStore, logStore, cfg.AgentToken)

	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	go func() {
		err := srv.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("Server shutdown complete")
		} else if err != nil {
			logger.Error("Server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	logger.Info("Server started", slog.String("port", cfg.Port), slog.String("env", Environment))
	<-killSig

	logger.Info("Shutting down server")

	logger.Info("scheduleStore")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}

// runScheduler dispatches due update tasks and fails stale running ones.
func runScheduler(deviceStore *dbstore.DeviceStore, taskStore *dbstore.TaskStore, logStore *dbstore.LogStore, agentToken string) {
	client := agent.NewClient(agentToken)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		scheduler.Tick(deviceStore, taskStore, logStore, client, time.Now())
	}
}
