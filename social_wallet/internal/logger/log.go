package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New() *slog.Logger {
	format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT"))) // "json" | "text"
	level := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))   // "debug" | "info" | "warn" | "error"

	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl, AddSource: envBool("LOG_SOURCE", false)}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

func envBool(k string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(k))) {
	case "1", "true", "yes", "y":
		return true
	case "0", "false", "no", "n":
		return false
	}
	return def
}
