package main

import (
	"log/slog"
	"os"
)

// initLogging は stderr へ JSON またはテキストで slog を出す。
// Cloud Run では K_SERVICE が付くため JSON をデフォルトにし、ローカルでテキストにしたい場合は LOG_FORMAT=text。
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
	// Cloud Run / GKE などでは JSON のほうが Cloud Logging で扱いやすい
	return os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""
}
