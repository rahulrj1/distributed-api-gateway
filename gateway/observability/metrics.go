package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts total requests by service, method, and status
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests processed by the gateway",
		},
		[]string{"service", "method", "status"},
	)

	// RequestDuration tracks request latency
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method"},
	)

	// RateLimitRejections counts rate limit rejections
	RateLimitRejections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_rejections_total",
			Help: "Total number of requests rejected due to rate limiting",
		},
		[]string{"client_id"},
	)

	// CircuitBreakerState tracks circuit breaker state per service (0=closed, 1=open, 2=half-open)
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service"},
	)
)

// Circuit breaker state values
const (
	CircuitClosed   = 0
	CircuitOpen     = 1
	CircuitHalfOpen = 2
)
