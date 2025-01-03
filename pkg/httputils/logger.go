package httputils

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

type contextKey string

const LogRequestIdKey = contextKey("request_id")
const LogUserIdKey = contextKey("user_id")

type idHandler struct {
	slog.Handler
}

func (h *idHandler) Handle(ctx context.Context, record slog.Record) error {
	id, ok := ctx.Value(LogRequestIdKey).(string)
	if ok {
		record.AddAttrs(slog.String("request_id", id))
	}
	id, ok = ctx.Value(LogUserIdKey).(string)
	if ok {
		record.AddAttrs(slog.String("user_id", id))
	}
	return h.Handler.Handle(ctx, record)
}

// slogErrorWriter adapts slog.Logger to io.Writer so that it can passed to http.Server struct ErrorLog
type slogErrorWriter struct {
	logger *slog.Logger
}

func NewSlogErrorWriter(logger *slog.Logger) *log.Logger {
	return log.New(&slogErrorWriter{logger: logger}, "", 0)
}

func (w *slogErrorWriter) Write(p []byte) (n int, err error) {
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
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time in UTC
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.UTC().Format(time.RFC3339))
				}
			}
			return a
		},
		Level: getLogLevelFromString(level),
	})}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
