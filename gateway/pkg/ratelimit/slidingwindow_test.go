package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/distributed-api-gateway/gateway/pkg/redis"
)

func TestLimiterAllow(t *testing.T) {
	// Skip if no Redis available
	client := redis.New("localhost:6379")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	limiter := NewLimiter(client, 5) // 5 requests per minute

	// Clear any existing keys
	testKey := "test-client-" + time.Now().Format("150405")

	t.Run("allows requests under limit", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			result := limiter.Allow(context.Background(), testKey)
			if !result.Allowed {
				t.Errorf("Request %d should be allowed", i+1)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		result := limiter.Allow(context.Background(), testKey)
		if result.Allowed {
			t.Error("Request should be blocked after exceeding limit")
		}
		if result.RetryAfter <= 0 {
			t.Error("RetryAfter should be positive when rate limited")
		}
	})
}

func TestLimiterFailOpen(t *testing.T) {
	// Use invalid Redis address
	client := redis.New("invalid:9999")
	limiter := NewLimiter(client, 5)

	result := limiter.Allow(context.Background(), "test-key")
	if !result.Allowed {
		t.Error("Should fail-open when Redis is unavailable")
	}
}
