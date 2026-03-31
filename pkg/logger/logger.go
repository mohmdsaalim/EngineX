package logger

import (
	"log/slog"
	"os"
)

// New creates a JSON logger tagged with the service name.
// Usage: log := logger.New("gateway")
//        log.Info("server started", "port", ":8080")
func New(service string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(handler).With("service", service)
}