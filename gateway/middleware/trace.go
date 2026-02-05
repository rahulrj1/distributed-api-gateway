package middleware

import (
	"net/http"
	"time"

	"github.com/distributed-api-gateway/gateway/pkg/trace"
)

// Trace middleware initializes request tracing.
// It extracts or generates a trace ID and adds it to the request context.
func Trace(pub *trace.Publisher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate or extract trace ID
			traceID := trace.FromRequest(r)

			// Add trace ID to response headers for debugging
			w.Header().Set(trace.TraceHeader, traceID)

			// Add trace context
			ctx := trace.WithTraceID(r.Context(), traceID)
			ctx = trace.WithPublisher(ctx, pub)

			// Publish initial request event
			if pub.Enabled() {
				info := &trace.RequestInfo{
					TraceID:   traceID,
					Method:    r.Method,
					Path:      r.URL.Path,
					Service:   trace.ExtractService(r.URL.Path),
					Timestamp: time.Now(),
				}
				pub.PublishRequest(ctx, info)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
