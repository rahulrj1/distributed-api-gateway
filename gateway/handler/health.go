package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HealthHandler returns a handler for the /health endpoint
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	}
}

// MetricsHandler returns Prometheus metrics handler
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
