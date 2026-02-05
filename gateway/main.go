package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/distributed-api-gateway/gateway/config"
	"github.com/distributed-api-gateway/gateway/handler"
	"github.com/distributed-api-gateway/gateway/middleware"
	"github.com/distributed-api-gateway/gateway/pkg/jwt"
	"github.com/distributed-api-gateway/gateway/pkg/ratelimit"
	"github.com/distributed-api-gateway/gateway/pkg/redis"
	"github.com/distributed-api-gateway/gateway/pkg/trace"
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

	// Setup Redis and rate limiter
	redisClient := redis.New(config.DefaultRedisAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	err = redisClient.Ping(ctx)
	cancel()
	if err != nil {
		log.Printf("Redis unavailable, rate limiting will fail-open: %v", err)
	} else {
		log.Printf("Redis connected at %s", config.DefaultRedisAddr)
	}
	limiter := ratelimit.NewLimiter(redisClient, config.DefaultRateLimit)

	// Setup trace publisher for pipeline visualization
	tracePublisher := trace.NewPublisher(redisClient.Raw())
	log.Printf("Trace visualization enabled")

	// Create handlers and middleware chain: Trace → Metrics → Auth → RateLimit → Proxy
	forwarder := proxy.NewForwarder()
	proxyHandler := handler.ProxyHandler(routes, forwarder, redisClient)
	rateLimitMiddleware := middleware.RateLimit(limiter)
	authMiddleware := middleware.Auth(validator)
	metricsMiddleware := middleware.Metrics()
	traceMiddleware := middleware.Trace(tracePublisher)
	log.Printf("Circuit breaker enabled")

	// Create router
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthHandler())
	mux.Handle("/metrics", handler.MetricsHandler())
	mux.HandleFunc("/ws/trace/", handler.TraceWebSocket(redisClient.Raw()))
	mux.Handle("/", traceMiddleware(metricsMiddleware(authMiddleware(rateLimitMiddleware(proxyHandler)))))

	// Start server with CORS support for visualizer
	log.Printf("Starting gateway on %s", cfg.Address())
	if err := http.ListenAndServe(cfg.Address(), middleware.CORS(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
