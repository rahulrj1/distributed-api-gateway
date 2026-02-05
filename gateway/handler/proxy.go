package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/distributed-api-gateway/gateway/observability"
	"github.com/distributed-api-gateway/gateway/pkg/circuitbreaker"
	"github.com/distributed-api-gateway/gateway/pkg/redis"
	"github.com/distributed-api-gateway/gateway/pkg/trace"
	"github.com/distributed-api-gateway/gateway/proxy"
)

// ErrorResponse represents the standard error response format per LLD
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ProxyHandler creates a handler for proxying requests to backend services
func ProxyHandler(routes *config.RoutesConfig, forwarder *proxy.Forwarder, redisClient *redis.Client) http.HandlerFunc {
	breakers := make(map[string]*circuitbreaker.Breaker)

	return func(w http.ResponseWriter, r *http.Request) {
		// Match route
		route := routes.MatchRoute(r.URL.Path)
		if route == nil {
			writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Route not found")
			return
		}

		// Get or create circuit breaker for this service
		service := route.PathPrefix // Use path prefix as service identifier
		breaker, ok := breakers[service]
		if !ok {
			breaker = circuitbreaker.NewBreaker(redisClient, service)
			breakers[service] = breaker
		}

		// Check circuit state and record metric
		cbStart := time.Now()
		cbResult := breaker.Allow(r.Context())
		updateCircuitBreakerMetric(service, cbResult.State)
		if !cbResult.Allowed {
			trace.EmitStep(r.Context(), trace.StepCircuit, trace.StatusFailed, time.Since(cbStart), map[string]interface{}{
				"service": service,
				"state":   string(cbResult.State),
				"reason":  "circuit open",
			})
			writeError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "circuit breaker open")
			return
		}
		trace.EmitStep(r.Context(), trace.StepCircuit, trace.StatusSuccess, time.Since(cbStart), map[string]interface{}{
			"service": service,
			"state":   string(cbResult.State),
		})

		// Forward request
		fwdStart := time.Now()
		if err := forwarder.Forward(w, r, route); err != nil {
			breaker.RecordFailure(r.Context()) // Record failure for circuit breaker
			if proxyErr, ok := err.(*proxy.ProxyError); ok {
				code := "BAD_GATEWAY"
				if proxyErr.Code == http.StatusGatewayTimeout {
					code = "GATEWAY_TIMEOUT"
				}
				trace.EmitStep(r.Context(), trace.StepForward, trace.StatusFailed, time.Since(fwdStart), map[string]interface{}{
					"service":     service,
					"target":      route.Target,
					"error":       proxyErr.Message,
					"status_code": proxyErr.Code,
				})
				writeError(w, r, proxyErr.Code, code, proxyErr.Message)
				return
			}
			log.Printf("Unexpected proxy error: %v", err)
			trace.EmitStep(r.Context(), trace.StepForward, trace.StatusFailed, time.Since(fwdStart), map[string]interface{}{
				"service": service,
				"error":   "internal error",
			})
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}

		trace.EmitStep(r.Context(), trace.StepForward, trace.StatusSuccess, time.Since(fwdStart), map[string]interface{}{
			"service": service,
			"target":  route.Target,
		})

		// Emit complete event
		trace.EmitStep(r.Context(), trace.StepComplete, trace.StatusSuccess, 0, nil)

		breaker.RecordSuccess(r.Context()) // Record success
	}
}

// writeError writes a standard error response per LLD format
func writeError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: r.Header.Get("X-Request-ID"), // From client if provided
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func updateCircuitBreakerMetric(service string, state circuitbreaker.State) {
	var value float64
	switch state {
	case circuitbreaker.StateClosed:
		value = observability.CircuitClosed
	case circuitbreaker.StateOpen:
		value = observability.CircuitOpen
	case circuitbreaker.StateHalfOpen:
		value = observability.CircuitHalfOpen
	}
	observability.CircuitBreakerState.WithLabelValues(service).Set(value)
}
