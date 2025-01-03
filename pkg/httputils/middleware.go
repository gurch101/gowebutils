package httputils

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// LoggingMiddleware logs the request and response details.
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

// RecoveryMiddleware recovers from panics and sends a 500 Internal Server Error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				ServerErrorResponse(w, r, fmt.Errorf("%s", err))
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

func RateLimitMiddleware(next http.Handler) http.Handler {
	rateLimitConfig := &RateLimitConfig{rate: 10, burst: 20}
	rlEnabled := os.Getenv("RATE_LIMIT_ENABLED")
	rateLimitConfig.enabled = rlEnabled == "" || rlEnabled == "true"
	if !rateLimitConfig.enabled {
		return next
	}
	rateLimitRate := os.Getenv("RATE_LIMIT_RATE")
	if rateLimitRate != "" {
		rateLimitRate, err := strconv.ParseFloat(rateLimitRate, 64)

		if err != nil {
			panic(err)
		}

		rateLimitConfig.rate = rateLimitRate
	}

	burst := os.Getenv("RATE_LIMIT_BURST")
	if burst != "" {
		burst, err := strconv.Atoi(burst)

		if err != nil {
			panic(err)
		}

		rateLimitConfig.burst = burst
	}

	slog.Info("Rate limit middleware enabled", "rate", rateLimitConfig.rate, "burst", rateLimitConfig.burst)

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
			ServerErrorResponse(w, r, fmt.Errorf("could not parse remote address: %v", err))
		}

		mu.Lock()
		if _, ok := clients[ip]; !ok {
			// Rate limit to 2 requests per second with a burst of 4 requests. 4 is the bucket size, 2 is the refill rate/s.
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(rateLimitConfig.rate), rateLimitConfig.burst)}
		}

		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			RateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
