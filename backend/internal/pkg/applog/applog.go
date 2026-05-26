// Package applog provides structured application logging via slog.
package applog

import (
	"log/slog"
	"os"
)

// Init configures JSON slog output for production, text for development.
func Init(environment string) {
	level := slog.LevelInfo
	if environment != "production" {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
