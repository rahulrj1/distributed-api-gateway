package main

import (
	"log"
	"net/http"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/distributed-api-gateway/gateway/handler"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create router
	mux := http.NewServeMux()

	// Register health endpoints
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.HandleFunc("/metrics", handler.MetricsHandler())

	// Start server
	addr := cfg.Address()
	log.Printf("Starting gateway server on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
