package main

import (
	"log"
	"net/http"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/distributed-api-gateway/gateway/handler"
	"github.com/distributed-api-gateway/gateway/proxy"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Load routes
	routes, err := config.LoadRoutes(config.DefaultRoutesPath)
	if err != nil {
		log.Fatalf("Failed to load routes: %v", err)
	}
	log.Printf("Loaded %d routes", len(routes.Routes))

	// Create forwarder
	forwarder := proxy.NewForwarder()

	// Create router
	mux := http.NewServeMux()

	// Register health endpoints (these don't go through proxy)
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.HandleFunc("/metrics", handler.MetricsHandler())

	// Register proxy handler for all other routes
	mux.HandleFunc("/", handler.ProxyHandler(routes, forwarder))

	// Start server
	addr := cfg.Address()
	log.Printf("Starting gateway server on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
