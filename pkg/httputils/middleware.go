package httputils

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestId := r.Header.Get("X-Request-ID")
		if requestId != "" {
			ctx := context.WithValue(r.Context(), LogRequestIdKey, fmt.Sprintf("ext-%s", requestId))
			r = r.WithContext(ctx)
		} else {
			id := uuid.New()
			ctx := context.WithValue(r.Context(), LogRequestIdKey, id.String())
			r = r.WithContext(ctx)
		}

		slog.InfoContext(r.Context(), "request started")

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		slog.InfoContext(r.Context(), "request completed", "request_method", r.Method, "request_url", r.URL.String(), "duration", duration.Milliseconds())
	})
}
