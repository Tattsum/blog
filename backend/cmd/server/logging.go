package main

import (
	"log/slog"
	"os"
)

func initLogging() {
	var h slog.Handler
	if useJSONLog() {
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}

func useJSONLog() bool {
	if os.Getenv("LOG_FORMAT") == "text" {
		return false
	}
	if os.Getenv("LOG_FORMAT") == "json" {
		return true
	}
	return os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""
}
