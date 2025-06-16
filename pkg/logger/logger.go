package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional convenience methods
type Logger struct {
	*slog.Logger
}

// New creates a new structured logger
func New(level string) *Logger {
	// Parse log level
	var logLevel slog.Level
	switch level {
	case "debug", "DEBUG":
		logLevel = slog.LevelDebug
	case "info", "INFO":
		logLevel = slog.LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		logLevel = slog.LevelWarn
	case "error", "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create handler with options
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	// Use JSON handler for production, text handler for development
	var handler slog.Handler
	if os.Getenv("ENVIRONMENT") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// RequestLogger creates a logger with request context
func (l *Logger) RequestLogger(requestID, method, path string) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
		),
	}
}

// WithComponent creates a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("component", component)),
	}
}

// WithError logs an error with additional context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("error", err.Error())),
	}
}

// DatabaseError logs database operation errors
func (l *Logger) DatabaseError(ctx context.Context, operation string, table string, err error) {
	l.Logger.ErrorContext(ctx, "Database operation failed",
		slog.String("operation", operation),
		slog.String("table", table),
		slog.String("error", err.Error()),
	)
}

// APIError logs API errors with request context
func (l *Logger) APIError(ctx context.Context, endpoint string, statusCode int, err error) {
	l.Logger.ErrorContext(ctx, "API error",
		slog.String("endpoint", endpoint),
		slog.Int("status_code", statusCode),
		slog.String("error", err.Error()),
	)
}

// APIRequest logs incoming API requests
func (l *Logger) APIRequest(ctx context.Context, method, path, remoteAddr string) {
	l.Logger.InfoContext(ctx, "API request",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("remote_addr", remoteAddr),
	)
}

// APIResponse logs API responses
func (l *Logger) APIResponse(ctx context.Context, method, path string, statusCode int, duration string) {
	l.Logger.InfoContext(ctx, "API response",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status_code", statusCode),
		slog.String("duration", duration),
	)
}