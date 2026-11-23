package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"goth/internal/agent"
)

func envOr(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func heartbeatInterval() time.Duration {
	seconds, err := strconv.Atoi(envOr("AGENT_HEARTBEAT_SECS", "30"))
	if err != nil || seconds < 5 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}

func heartbeatLoop(handler *agent.StatusHandler, serverURL string, agentURL string, token string) {
	if serverURL == "" {
		log.Println("AGENT_SERVER not set, auto-registration disabled")
		return
	}

	beat := func() {
		if !handler.IsPowered() {
			log.Println("powered off, skipping heartbeat")
			return
		}
		status := handler.Snapshot()
		err := agent.SendRegistration(serverURL, agent.Registration{
			Name:          status.Name,
			Version:       status.Version,
			OS:            status.OS,
			UptimeSeconds: status.UptimeSeconds,
			URL:           agentURL,
			Token:         token,
		})
		if err != nil {
			log.Println("heartbeat failed:", err)
			return
		}
		log.Println("registered with", serverURL)
	}

	beat()

	ticker := time.NewTicker(heartbeatInterval())
	defer ticker.Stop()
	for range ticker.C {
		beat()
	}
}

func main() {
	port := envOr("PORT", ":4100")
	token := os.Getenv("AGENT_TOKEN")

	statusHandler := agent.NewStatusHandler(
		envOr("AGENT_NAME", "office-pc"),
		envOr("AGENT_VERSION", "1.0.0"),
		envOr("AGENT_OS", "NixOS 25.05"),
		token,
	)
	firewallHandler := agent.NewFirewallHandler(agent.NewRuleStore(agent.DefaultFirewallRules()), token)
	updateHandler := agent.NewUpdateHandler(statusHandler, token)
	powerHandler := agent.NewPowerHandler(statusHandler, token)
	consoleHandler := agent.NewConsoleHandler(statusHandler, token)

	http.Handle("/status", statusHandler)
	http.Handle("/firewall", firewallHandler)
	http.Handle("/firewall/apply", firewallHandler)
	http.Handle("/update", updateHandler)
	http.Handle("/power", powerHandler)
	http.Handle("/console", consoleHandler)

	go heartbeatLoop(statusHandler, os.Getenv("AGENT_SERVER"), envOr("AGENT_URL", "http://localhost"+port), token)

	log.Println("agent listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
