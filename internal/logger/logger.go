package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

const (
	DefaultLoggerFileName = "logs.log"
	DailyChanne           = "daily"
	StackChanne           = "stack"
)

var Log *slog.Logger

func Init(destination io.Writer, l string) {
	Log = slog.New(slog.NewJSONHandler(destination, &slog.HandlerOptions{
		Level: getLevel(l),
	}))
}

func LogError(ctx context.Context, err error, message string) {
	if message == "" {
		message = "something went wrong"
	}

	if Log != nil {
		Log.ErrorContext(ctx, message, slog.String("err", err.Error()))
	}
}

func getLevel(l string) slog.Level {
	switch {
	case strings.ToLower(l) == "info":
		return slog.LevelInfo
	case strings.ToLower(l) == "warn":
		return slog.LevelWarn
	case strings.ToLower(l) == "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
