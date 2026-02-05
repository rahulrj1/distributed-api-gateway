package circuitbreaker

import (
	"context"
	"testing"
	"time"

	"github.com/distributed-api-gateway/gateway/pkg/redis"
)

func getTestClient(t *testing.T) *redis.Client {
	client := redis.New("localhost:6379")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	return client
}

func TestBreakerInitialState(t *testing.T) {
	client := getTestClient(t)
	breaker := NewBreaker(client, "test-init-"+time.Now().Format("150405.000"))

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Circuit should be closed initially")
	}
	if result.State != StateClosed {
		t.Errorf("Expected CLOSED, got %s", result.State)
	}
}

func TestBreakerFailOpen(t *testing.T) {
	client := redis.New("invalid:9999")
	breaker := NewBreaker(client, "test-service")

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Should fail-open when Redis unavailable")
	}
}

func TestBreakerOpensAfterFailures(t *testing.T) {
	client := getTestClient(t)
	breaker := NewBreaker(client, "test-open-"+time.Now().Format("150405.000"))

	// Record 10 requests: all failures (100% failure rate, > threshold of 5)
	for i := 0; i < 10; i++ {
		breaker.Allow(context.Background())
		breaker.RecordFailure(context.Background())
	}

	// Circuit should now be OPEN
	result := breaker.Allow(context.Background())
	if result.Allowed {
		t.Error("Circuit should be OPEN and deny requests")
	}
	if result.State != StateOpen {
		t.Errorf("Expected OPEN, got %s", result.State)
	}
}

func TestBreakerStaysClosedWithLowFailureRate(t *testing.T) {
	client := getTestClient(t)
	breaker := NewBreaker(client, "test-lowrate-"+time.Now().Format("150405.000"))

	// Record 10 requests: 3 failures, 7 successes (30% failure rate < 50% threshold)
	for i := 0; i < 10; i++ {
		breaker.Allow(context.Background())
		if i < 3 {
			breaker.RecordFailure(context.Background())
		} else {
			breaker.RecordSuccess(context.Background())
		}
	}

	// Circuit should still be CLOSED
	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Circuit should remain CLOSED with low failure rate")
	}
}

func TestBreakerSuccessesKeepClosed(t *testing.T) {
	client := getTestClient(t)
	breaker := NewBreaker(client, "test-success-"+time.Now().Format("150405.000"))

	for i := 0; i < 20; i++ {
		breaker.Allow(context.Background())
		breaker.RecordSuccess(context.Background())
	}

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Circuit should remain CLOSED with only successes")
	}
}
