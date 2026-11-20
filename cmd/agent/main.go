package main

import (
	"log"
	"net/http"
	"os"

	"goth/internal/agent"
)

func envOr(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	statusHandler := agent.NewStatusHandler(
		envOr("AGENT_NAME", "office-pc"),
		envOr("AGENT_VERSION", "1.0.0"),
		envOr("AGENT_OS", "NixOS 25.05"),
	)

	http.Handle("/status", statusHandler)

	port := envOr("PORT", ":4100")
	log.Println("agent listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
