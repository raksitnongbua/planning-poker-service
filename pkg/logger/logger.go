package logger

import (
	"io"
	"log/slog"
	"os"
)

var Logger *slog.Logger

func Init(appEnv string) {
	if appEnv == "local" || appEnv == "development" {
		Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		Logger.Info("logger initialized", "env", appEnv)
	} else {
		Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}
