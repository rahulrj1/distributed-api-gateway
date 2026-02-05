package middleware

import (
	"net/http"
	"strconv"

	"github.com/distributed-api-gateway/gateway/observability"
	"github.com/distributed-api-gateway/gateway/pkg/ratelimit"
)

// RateLimit returns middleware that limits requests per client.
// Uses X-Client-ID from JWT, falls back to X-Forwarded-For or RemoteAddr.
func RateLimit(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := getRateLimitKey(r)
			result := limiter.Allow(r.Context(), key)

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(ratelimit.DefaultLimit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, ratelimit.DefaultLimit-result.Count)))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
				observability.RateLimitRejections.WithLabelValues(key).Inc()
				writeRateLimitError(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getRateLimitKey returns the key for rate limiting: client_id > X-Forwarded-For > RemoteAddr
func getRateLimitKey(r *http.Request) string {
	if clientID := r.Header.Get("X-Client-ID"); clientID != "" {
		return clientID
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}

func writeRateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"rate limit exceeded"}}`))
}
