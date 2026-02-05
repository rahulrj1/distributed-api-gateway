package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/distributed-api-gateway/gateway/observability"
	"github.com/distributed-api-gateway/gateway/pkg/ratelimit"
	"github.com/distributed-api-gateway/gateway/pkg/trace"
)

// RateLimit returns middleware that limits requests per client.
// Uses X-Client-ID from JWT, falls back to X-Forwarded-For or RemoteAddr.
func RateLimit(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			key := getRateLimitKey(r)
			result := limiter.Allow(r.Context(), key)

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(ratelimit.DefaultLimit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, ratelimit.DefaultLimit-result.Count)))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(result.RetryAfter))
				observability.RateLimitRejections.WithLabelValues(key).Inc()

				trace.EmitStep(r.Context(), trace.StepRateLimit, trace.StatusFailed, time.Since(start), map[string]interface{}{
					"client_id":   key,
					"count":       result.Count,
					"limit":       ratelimit.DefaultLimit,
					"retry_after": result.RetryAfter,
				})
				writeRateLimitError(w)
				return
			}

			trace.EmitStep(r.Context(), trace.StepRateLimit, trace.StatusSuccess, time.Since(start), map[string]interface{}{
				"client_id": key,
				"count":     result.Count,
				"remaining": ratelimit.DefaultLimit - result.Count,
			})

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
