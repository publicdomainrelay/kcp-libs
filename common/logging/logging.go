package logging

import (
	"io"
	"log/slog"
)

type Options struct {
	Service string

	Writer io.Writer

	Level slog.Level
}

func New(opts Options) *slog.Logger {
	if opts.Level == 0 {
		opts.Level = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(opts.Writer, &slog.HandlerOptions{Level: opts.Level})
	logger := slog.New(handler)
	if opts.Service != "" {
		logger = logger.With("service", opts.Service)
	}
	return logger
}

func Discard() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}
