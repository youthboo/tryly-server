package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

func Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	lower := strings.ToLower(message)
	level := slog.LevelInfo
	switch {
	case strings.Contains(lower, "[error]") || strings.Contains(lower, " error"):
		level = slog.LevelError
	case strings.Contains(lower, "[warn]") || strings.Contains(lower, " warn"):
		level = slog.LevelWarn
	case strings.Contains(lower, "[debug]"):
		level = slog.LevelDebug
	}
	slog.Log(context.Background(), level, "app log", "message", message)
}
