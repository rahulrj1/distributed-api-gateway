package main

import (
	"log"
	"net/http"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/distributed-api-gateway/gateway/handler"
	"github.com/distributed-api-gateway/gateway/middleware"
	"github.com/distributed-api-gateway/gateway/pkg/jwt"
	"github.com/distributed-api-gateway/gateway/proxy"
)

func main() {
	cfg := config.Load()

	// Load routes
	routes, err := config.LoadRoutes(config.DefaultRoutesPath)
	if err != nil {
		log.Fatalf("Failed to load routes: %v", err)
	}
	log.Printf("Loaded %d routes", len(routes.Routes))

	// Setup JWT validator
	validator, err := jwt.NewValidator(config.DefaultPublicKeyPath, "")
	if err != nil {
		log.Fatalf("Failed to load JWT public key: %v", err)
	}
	log.Printf("JWT auth enabled")

	// Create handlers
	forwarder := proxy.NewForwarder()
	proxyHandler := handler.ProxyHandler(routes, forwarder)
	authMiddleware := middleware.Auth(validator)

	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.HandleFunc("/metrics", handler.MetricsHandler())
	mux.Handle("/", authMiddleware(proxyHandler))

	// Start server
	log.Printf("Starting gateway on %s", cfg.Address())
	if err := http.ListenAndServe(cfg.Address(), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
