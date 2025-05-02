package httputils

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const (
	LogUserIDKey = contextKey("user_id")
)

type idHandler struct {
	slog.Handler
}

func (h *idHandler) Handle(ctx context.Context, record slog.Record) error {
	id, ok := ctx.Value(middleware.RequestIDKey).(string)
	if ok {
		record.AddAttrs(slog.String("request_id", id))
	}

	id, ok = ctx.Value(LogUserIDKey).(string)
	if ok {
		record.AddAttrs(slog.String("user_id", id))
	}

	err := h.Handler.Handle(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to log record: %w", err)
	}

	return nil
}

// slogErrorWriter adapts slog.Logger to io.Writer so that it can passed to http.Server struct ErrorLog.
type slogErrorWriter struct {
	logger *slog.Logger
}

func NewSlogErrorWriter(logger *slog.Logger) *log.Logger {
	return log.New(&slogErrorWriter{logger: logger}, "", 0)
}

func (w *slogErrorWriter) Write(p []byte) (int, error) {
	// Log the message using slog.Logger
	w.logger.Error(string(p))

	return len(p), nil
}

func getLogLevelFromString(level string) slog.Level {
	if level == "" {
		return slog.LevelInfo
	}

	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func InitializeSlog(level string) *slog.Logger {
	handler := &idHandler{slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			// Format time in UTC
			if attr.Key == slog.TimeKey {
				if t, ok := attr.Value.Any().(time.Time); ok {
					attr.Value = slog.StringValue(t.UTC().Format(time.RFC3339))
				}
			}

			return attr
		},
		Level: getLogLevelFromString(level),
	})}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

type SlogLogFormatter struct {
	Logger *slog.Logger
}

// NewSlogLogFormatter creates a new SlogLogFormatter.
func NewSlogLogFormatter(logger *slog.Logger) *SlogLogFormatter {
	return &SlogLogFormatter{Logger: logger}
}

// NewLogEntry creates a new LogEntry for the request.
func (f *SlogLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry { //nolint:ireturn
	return &SlogLogEntry{
		logger:  f.Logger,
		request: r,
	}
}

type SlogLogEntry struct {
	logger  *slog.Logger
	request *http.Request
}

// Write logs the request completion details.
func (e *SlogLogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ interface{}) {
	id, ok := e.request.Context().Value(middleware.RequestIDKey).(string)
	requestID := "-"

	if ok {
		requestID = id
	}

	e.logger.InfoContext(e.request.Context(), "request completed",
		slog.String("method", e.request.Method),
		slog.String("path", e.request.URL.Path),
		slog.Int("status", status),
		slog.Int("bytes", bytes),
		slog.Duration("elapsed", elapsed),
		slog.String("ip", e.request.RemoteAddr),
		slog.String("request_id", requestID),
	)
}

// Panic logs a panic with its stack trace.
func (e *SlogLogEntry) Panic(v interface{}, stack []byte) {
	e.logger.ErrorContext(e.request.Context(), "request panicked",
		slog.String("method", e.request.Method),
		slog.String("path", e.request.URL.Path),
		slog.String("ip", e.request.RemoteAddr),
		slog.Any("panic", v),
		slog.String("stack", string(stack)),
	)
}
