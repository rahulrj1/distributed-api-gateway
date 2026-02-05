package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	traceIDKey    contextKey = "trace_id"
	publisherKey  contextKey = "trace_publisher"
	requestKey    contextKey = "trace_request"
	stepTimerKey  contextKey = "trace_step_timer"
)

// TraceHeader is the HTTP header used to pass trace IDs.
const TraceHeader = "X-Trace-ID"

// GenerateID creates a new unique trace ID.
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// FromRequest extracts or generates a trace ID from the request.
func FromRequest(r *http.Request) string {
	if id := r.Header.Get(TraceHeader); id != "" {
		return id
	}
	return GenerateID()
}

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID retrieves the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// WithPublisher adds a trace publisher to the context.
func WithPublisher(ctx context.Context, pub *Publisher) context.Context {
	return context.WithValue(ctx, publisherKey, pub)
}

// GetPublisher retrieves the trace publisher from the context.
func GetPublisher(ctx context.Context) *Publisher {
	if pub, ok := ctx.Value(publisherKey).(*Publisher); ok {
		return pub
	}
	return nil
}

// StartStep marks the beginning of a pipeline step and returns a function to complete it.
func StartStep(ctx context.Context, step Step) func(status Status, err string) {
	pub := GetPublisher(ctx)
	traceID := GetTraceID(ctx)
	start := time.Now()

	// Emit "running" event
	if pub != nil && traceID != "" {
		event := NewEvent(traceID, step, StatusRunning)
		pub.Publish(ctx, event)
	}

	// Return completion function
	return func(status Status, errMsg string) {
		if pub == nil || traceID == "" {
			return
		}
		event := NewEvent(traceID, step, status).WithDuration(time.Since(start))
		if errMsg != "" {
			event.WithError(errMsg)
		}
		pub.Publish(ctx, event)
	}
}

// EmitStep is a convenience function to emit a single step event.
func EmitStep(ctx context.Context, step Step, status Status, duration time.Duration, details map[string]interface{}) {
	pub := GetPublisher(ctx)
	traceID := GetTraceID(ctx)
	if pub == nil || traceID == "" {
		return
	}

	event := NewEvent(traceID, step, status).WithDuration(duration)
	for k, v := range details {
		event.WithDetail(k, v)
	}
	pub.Publish(ctx, event)
}

// ExtractService extracts the service name from a request path.
// Example: /service-a/hello -> service-a
func ExtractService(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
