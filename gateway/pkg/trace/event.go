// Package trace provides request tracing for pipeline visualization.
package trace

import (
	"encoding/json"
	"time"
)

// Step represents a stage in the request pipeline.
type Step string

const (
	StepReceived     Step = "RECEIVED"      // Request received by gateway
	StepAuth         Step = "AUTH"          // JWT authentication
	StepRateLimit    Step = "RATE_LIMIT"    // Rate limiting check
	StepCircuit      Step = "CIRCUIT"       // Circuit breaker check
	StepForward      Step = "FORWARD"       // Forwarding to backend
	StepResponse     Step = "RESPONSE"      // Response received from backend
	StepComplete     Step = "COMPLETE"      // Request completed
)

// Status represents the outcome of a pipeline step.
type Status string

const (
	StatusPending Status = "PENDING"  // Step not yet processed
	StatusRunning Status = "RUNNING"  // Step currently executing
	StatusSuccess Status = "SUCCESS"  // Step completed successfully
	StatusFailed  Status = "FAILED"   // Step failed
	StatusSkipped Status = "SKIPPED"  // Step was skipped
)

// Event represents a trace event emitted during request processing.
type Event struct {
	TraceID   string                 `json:"trace_id"`
	Step      Step                   `json:"step"`
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  int64                  `json:"duration_us,omitempty"` // Microseconds
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// RequestInfo contains metadata about the traced request.
type RequestInfo struct {
	TraceID   string            `json:"trace_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Service   string            `json:"service"`
	ClientID  string            `json:"client_id,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewEvent creates a new trace event.
func NewEvent(traceID string, step Step, status Status) *Event {
	return &Event{
		TraceID:   traceID,
		Step:      step,
		Status:    status,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// WithDuration sets the duration of the event.
func (e *Event) WithDuration(d time.Duration) *Event {
	e.Duration = d.Microseconds()
	return e
}

// WithError sets the error message.
func (e *Event) WithError(err string) *Event {
	e.Error = err
	return e
}

// WithDetail adds a detail to the event.
func (e *Event) WithDetail(key string, value interface{}) *Event {
	e.Details[key] = value
	return e
}

// JSON serializes the event to JSON.
func (e *Event) JSON() ([]byte, error) {
	return json.Marshal(e)
}
