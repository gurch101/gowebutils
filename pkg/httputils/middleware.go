package httputils

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gurch101/gowebutils/pkg/parser"
	"golang.org/x/time/rate"
)

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

// RateLimitMiddleware is a middleware that limits the number of requests per second per IP address.
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
		mutex   sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mutex.Lock()
			for ipAddress, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ipAddress)
				}
			}
			mutex.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ServerErrorResponse(w, r, fmt.Errorf("could not parse remote address: %w", err))
		}

		mutex.Lock()
		if _, ok := clients[ipAddress]; !ok {
			limiter := rate.NewLimiter(
				rate.Limit(rateLimitConfig.rate),
				rateLimitConfig.burst,
			)
			clients[ipAddress] = &client{limiter: limiter, lastSeen: time.Now()}
		} else {
			clients[ipAddress].lastSeen = time.Now()
		}

		if !clients[ipAddress].limiter.Allow() {
			mutex.Unlock()
			RateLimitExceededResponse(w, r)

			return
		}

		mutex.Unlock()

		next.ServeHTTP(w, r)
	})
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
