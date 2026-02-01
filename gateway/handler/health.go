package handler

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// MetricsResponse represents the metrics response (placeholder)
type MetricsResponse struct {
	// Placeholder for future metrics
}

// HealthHandler returns a handler for the /health endpoint
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := HealthResponse{
			Status: "healthy",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// MetricsHandler returns a handler for the /metrics endpoint (placeholder)
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Placeholder - will be implemented in Task 10 (Observability)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics placeholder\n"))
	}
}
