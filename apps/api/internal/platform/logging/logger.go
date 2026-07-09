package logging

import (
	"io"
	"log/slog"
	"strings"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

func New(cfg config.LogConfig, writer io.Writer) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		Level: level(cfg.Level),
	}

	if writer == nil {
		writer = io.Discard
	}

	if cfg.Format == config.LogFormatJSON {
		return slog.New(slog.NewJSONHandler(writer, handlerOptions))
	}

	return slog.New(slog.NewTextHandler(writer, handlerOptions))
}

func level(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case config.LogLevelDebug:
		return slog.LevelDebug
	case config.LogLevelWarn:
		return slog.LevelWarn
	case config.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
