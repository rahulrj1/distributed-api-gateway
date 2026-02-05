package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/distributed-api-gateway/gateway/observability"
)

// Metrics returns middleware that records request metrics
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			// Extract service from path (e.g., "/service-a/hello" → "/service-a")
			service := extractService(r.URL.Path)
			duration := time.Since(start).Seconds()

			// Record metrics
			observability.RequestsTotal.WithLabelValues(service, r.Method, strconv.Itoa(rw.statusCode)).Inc()
			observability.RequestDuration.WithLabelValues(service, r.Method).Observe(duration)
		})
	}
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// extractService gets service name from path: "/service-a/hello" → "/service-a"
func extractService(path string) string {
	if len(path) < 2 {
		return "unknown"
	}
	// Find second slash
	for i := 1; i < len(path); i++ {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return path
}
