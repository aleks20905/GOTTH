package handlers

import (
	"goth/internal/store"
	"time"
)

func writeLog(s store.LogStore, level string, category string, message string, details string, user string) {
	_ = s.AddEntry(&store.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Details:   details,
		UserEmail: user,
	})
}
