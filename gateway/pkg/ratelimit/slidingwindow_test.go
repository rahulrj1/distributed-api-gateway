package ratelimit

import (
	"context"
	"errors"
	"testing"
)

// mockRedis simulates Redis Eval responses for unit testing
type mockRedis struct {
	responses []interface{} // Queue of responses to return
	err       error         // Error to return (simulates Redis failure)
	calls     int           // Track number of calls
}

func (m *mockRedis) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.calls < len(m.responses) {
		resp := m.responses[m.calls]
		m.calls++
		return resp, nil
	}
	return []interface{}{int64(1), int64(0), int64(0)}, nil
}

func TestLimiterAllowsUnderLimit(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			[]interface{}{int64(1), int64(1), int64(0)}, // allowed=1, count=1, retry=0
			[]interface{}{int64(1), int64(2), int64(0)}, // allowed=1, count=2, retry=0
			[]interface{}{int64(1), int64(3), int64(0)}, // allowed=1, count=3, retry=0
		},
	}
	limiter := NewLimiter(mock, 5)

	for i := 0; i < 3; i++ {
		result := limiter.Allow(context.Background(), "test-client")
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
}

func TestLimiterBlocksOverLimit(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			[]interface{}{int64(0), int64(5), int64(30)}, // allowed=0, count=5, retry=30
		},
	}
	limiter := NewLimiter(mock, 5)

	result := limiter.Allow(context.Background(), "test-client")
	if result.Allowed {
		t.Error("Request should be blocked after exceeding limit")
	}
	if result.RetryAfter <= 0 {
		t.Error("RetryAfter should be positive when rate limited")
	}
}

func TestLimiterFailOpen(t *testing.T) {
	mock := &mockRedis{
		err: errors.New("connection refused"), // Simulate Redis failure
	}
	limiter := NewLimiter(mock, 5)

	result := limiter.Allow(context.Background(), "test-key")
	if !result.Allowed {
		t.Error("Should fail-open when Redis is unavailable")
	}
}

func TestLimiterInvalidResponse(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			"invalid", // Not an array - should fail-open
		},
	}
	limiter := NewLimiter(mock, 5)

	result := limiter.Allow(context.Background(), "test-key")
	if !result.Allowed {
		t.Error("Should fail-open on invalid response")
	}
}

func TestLimiterDefaultLimit(t *testing.T) {
	mock := &mockRedis{}
	limiter := NewLimiter(mock, 0) // 0 should use default

	if limiter.limit != DefaultLimit {
		t.Errorf("Expected default limit %d, got %d", DefaultLimit, limiter.limit)
	}
}

func TestLimiterRetryAfter(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			[]interface{}{int64(0), int64(100), int64(45)}, // blocked with 45s retry
		},
	}
	limiter := NewLimiter(mock, 100)

	result := limiter.Allow(context.Background(), "test-client")
	if result.RetryAfter != 45 {
		t.Errorf("Expected RetryAfter=45, got %d", result.RetryAfter)
	}
}
