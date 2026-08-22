// Package observability provides the structured logger and the canonical log
// field vocabulary shared by every layer of the backend (§122, §123).
package observability

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// LoggerOptions configures logger construction.
type LoggerOptions struct {
	// Level is one of debug, info, warn, error. Unknown values fall back to info.
	Level string
	// Format is json or text. Text is intended for local development only.
	Format string
}

// NewLogger builds the process logger.
func NewLogger(w io.Writer, opts LoggerOptions) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{Level: parseLevel(opts.Level)}

	var handler slog.Handler
	if strings.EqualFold(opts.Format, "text") {
		handler = slog.NewTextHandler(w, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(w, handlerOpts)
	}
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type loggerContextKey struct{}

// WithLogger returns a context carrying logger, so downstream layers inherit
// request-scoped attributes such as the request ID without threading a logger
// through every function signature.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// LoggerFrom returns the context logger, or the default logger when the
// context carries none.
func LoggerFrom(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}
