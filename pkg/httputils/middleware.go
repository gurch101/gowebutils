package httputils

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gurch101/gowebutils/pkg/parser"
	"golang.org/x/time/rate"
)

var ErrPanic = errors.New("panic")

// LoggingMiddleware logs the request and response details.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		RequestID := r.Header.Get("X-Request-ID")
		ctx := r.Context()

		if RequestID != "" {
			ctx = context.WithValue(ctx, LogRequestIDKey, "ext-"+RequestID)
		} else {
			id := uuid.New()
			ctx = context.WithValue(ctx, LogRequestIDKey, id.String())
		}

		slog.InfoContext(ctx, "request started")
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		slog.InfoContext(
			ctx,
			"request completed",
			"request_method", r.Method,
			"request_url", r.URL.String(),
			"duration", duration.Milliseconds(),
		)
	})
}

// RecoveryMiddleware recovers from panics and sends a 500 Internal Server Error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				ServerErrorResponse(w, r, fmt.Errorf("%w: %s", ErrPanic, err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type RateLimitConfig struct {
	enabled bool
	rate    float64
	burst   int
}

const (
	defaultRateLimitRate = 10

	defaultRateLimitBurst = 20
)

func getRateLimitConfig() *RateLimitConfig {
	rateLimitConfig := &RateLimitConfig{
		enabled: parser.ParseEnvBool("RATE_LIMIT_ENABLED", true),
		rate:    defaultRateLimitRate,
		burst:   defaultRateLimitBurst,
	}
	if !rateLimitConfig.enabled {
		return rateLimitConfig
	}

	rateLimit, err := parser.ParseEnvFloat64("RATE_LIMIT_RATE", rateLimitConfig.rate)
	if err != nil {
		panic(err)
	}

	rateLimitConfig.rate = rateLimit

	burst, err := parser.ParseEnvInt("RATE_LIMIT_BURST", rateLimitConfig.burst)
	if err != nil {
		panic(err)
	}

	rateLimitConfig.burst = burst

	return rateLimitConfig
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	rateLimitConfig := getRateLimitConfig()

	if !rateLimitConfig.enabled {
		return next
	}

	slog.Info("rate limit middleware enabled", "rate", rateLimitConfig.rate, "burst", rateLimitConfig.burst)

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ServerErrorResponse(w, r, fmt.Errorf("could not parse remote address: %w", err))
		}

		mu.Lock()
		if _, ok := clients[ip]; !ok {
			limiter := rate.NewLimiter(
				rate.Limit(rateLimitConfig.rate),
				rateLimitConfig.burst,
			)
			clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		} else {
			clients[ip].lastSeen = time.Now()
		}

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			RateLimitExceededResponse(w, r)

			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

type UnauthorizedRedirector func(w http.ResponseWriter, r *http.Request, destURL string)

func GetStateAwareAuthenticationMiddleware(_ UnauthorizedRedirector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func GetCORSMiddleware(trustedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add the "Vary: Origin" header.
			w.Header().Add("Vary", "Origin")
			// Get the value of the request's Origin header.
			origin := r.Header.Get("Origin")
			// Only run this if there's an Origin request header present.
			if origin != "" {
				// Loop through the list of trusted origins, checking to see if the request
				// origin exactly matches one of them. If there are no trusted origins, then
				// the loop won't be iterated.
				for i := range trustedOrigins {
					if origin == trustedOrigins[i] {
						// If there is a match, then set a "Access-Control-Allow-Origin"
						// response header with the request origin as the value and break
						// out of the loop.
						w.Header().Set("Access-Control-Allow-Origin", origin)

						break
					}
				}
			}

			// Call the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write writes the data to the gzip writer.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to write gzip response: %w", err)
	}

	return n, nil
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client supports gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Create a gzip writer
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			// Set the Content-Encoding header
			w.Header().Set("Content-Encoding", "gzip")

			// Wrap the response writer with the gzip writer
			gzResponseWriter := gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}
			next.ServeHTTP(gzResponseWriter, r)

			return
		}

		// If the client doesn't support gzip, serve the response as-is
		next.ServeHTTP(w, r)
	})
}
