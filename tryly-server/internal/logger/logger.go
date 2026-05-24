package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func Debug(message string, args ...interface{}) {
	slog.Debug(message, args...)
}

func Info(message string, args ...interface{}) {
	slog.Info(message, args...)
}

func Warn(message string, args ...interface{}) {
	slog.Warn(message, args...)
}

func Error(message string, args ...interface{}) {
	slog.Error(message, args...)
}

func Fatal(message string, args ...interface{}) {
	slog.Error(message, args...)
	os.Exit(1)
}

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
