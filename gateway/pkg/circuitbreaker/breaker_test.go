package circuitbreaker

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
	return []interface{}{int64(1), "CLOSED"}, nil
}

func TestBreakerInitialState(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			[]interface{}{int64(1), "CLOSED"}, // Allow returns: allowed=1, state=CLOSED
		},
	}
	breaker := NewBreaker(mock, "test-service")

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Circuit should be closed initially")
	}
	if result.State != StateClosed {
		t.Errorf("Expected CLOSED, got %s", result.State)
	}
}

func TestBreakerFailOpen(t *testing.T) {
	mock := &mockRedis{
		err: errors.New("connection refused"), // Simulate Redis failure
	}
	breaker := NewBreaker(mock, "test-service")

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Should fail-open when Redis unavailable")
	}
}

func TestBreakerOpensAfterFailures(t *testing.T) {
	// Build responses: 10x (Allow + RecordFailure) + final Allow
	// Each Allow returns CLOSED, RecordFailure returns 1, final Allow returns OPEN
	responses := make([]interface{}, 0, 21)
	for i := 0; i < 10; i++ {
		responses = append(responses, []interface{}{int64(1), "CLOSED"}) // Allow
		responses = append(responses, int64(1))                          // RecordFailure
	}
	responses = append(responses, []interface{}{int64(0), "OPEN"}) // Final Allow

	mock := &mockRedis{responses: responses}
	breaker := NewBreaker(mock, "test-service")

	// Simulate 10 requests with failures
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

func TestBreakerStaysClosedWithSuccesses(t *testing.T) {
	// 2x (Allow + RecordSuccess) + final Allow
	mock := &mockRedis{
		responses: []interface{}{
			[]interface{}{int64(1), "CLOSED"}, // Allow
			int64(1),                          // RecordSuccess
			[]interface{}{int64(1), "CLOSED"}, // Allow
			int64(1),                          // RecordSuccess
			[]interface{}{int64(1), "CLOSED"}, // Final Allow
		},
	}
	breaker := NewBreaker(mock, "test-service")

	// All successes - circuit stays closed
	for i := 0; i < 2; i++ {
		breaker.Allow(context.Background())
		breaker.RecordSuccess(context.Background())
	}

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Circuit should remain CLOSED with only successes")
	}
}

func TestBreakerHalfOpenState(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			[]interface{}{int64(1), "HALF_OPEN"}, // After cooldown, transitions to HALF_OPEN
		},
	}
	breaker := NewBreaker(mock, "test-service")

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("HALF_OPEN should allow requests")
	}
	if result.State != StateHalfOpen {
		t.Errorf("Expected HALF_OPEN, got %s", result.State)
	}
}

func TestBreakerInvalidResponse(t *testing.T) {
	mock := &mockRedis{
		responses: []interface{}{
			"invalid", // Not an array - should fail-open
		},
	}
	breaker := NewBreaker(mock, "test-service")

	result := breaker.Allow(context.Background())
	if !result.Allowed {
		t.Error("Should fail-open on invalid response")
	}
}
