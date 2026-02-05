package circuitbreaker

import (
	"context"
	"strconv"
	"time"

	"github.com/distributed-api-gateway/gateway/pkg/redis"
)

// State represents circuit breaker state
type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

// Config for circuit breaker
const (
	FailureThreshold   = 5  // Min failures to open circuit
	FailureRatePercent = 50 // Min failure rate % to open
	CooldownSeconds    = 30 // Time in OPEN before HALF_OPEN
	HalfOpenSuccesses  = 2  // Successes needed to close
	WindowSize         = 60 // Tracking window in seconds
)

// Lua script for circuit breaker state machine
// Keys: [state_key, window_key]
// Args: [now, cooldown, window_size, failure_threshold, failure_rate, half_open_successes]
// Returns: [allowed (0/1), state]
var circuitBreakerScript = `
local state_key = KEYS[1]
local window_key = KEYS[2]
local now = tonumber(ARGV[1])
local cooldown = tonumber(ARGV[2])
local window_size = tonumber(ARGV[3])
local failure_threshold = tonumber(ARGV[4])
local failure_rate = tonumber(ARGV[5])
local half_open_successes = tonumber(ARGV[6])

-- Get current state
local state_data = redis.call('HGETALL', state_key)
local state = 'CLOSED'
local opened_at = 0
local successes = 0

for i = 1, #state_data, 2 do
    if state_data[i] == 'state' then state = state_data[i+1] end
    if state_data[i] == 'opened_at' then opened_at = tonumber(state_data[i+1]) end
    if state_data[i] == 'successes' then successes = tonumber(state_data[i+1]) end
end

-- State machine
if state == 'OPEN' then
    if now - opened_at >= cooldown then
        -- Transition to HALF_OPEN
        redis.call('HSET', state_key, 'state', 'HALF_OPEN', 'successes', 0)
        return {1, 'HALF_OPEN'}
    end
    return {0, 'OPEN'}
end

if state == 'HALF_OPEN' then
    return {1, 'HALF_OPEN'}
end

-- CLOSED: check if we should open
local window = redis.call('HGETALL', window_key)
local total = 0
local failures = 0
for i = 1, #window, 2 do
    if window[i] == 'total' then total = tonumber(window[i+1]) end
    if window[i] == 'failures' then failures = tonumber(window[i+1]) end
end

if total >= failure_threshold then
    local rate = (failures / total) * 100
    if failures >= failure_threshold and rate >= failure_rate then
        redis.call('HSET', state_key, 'state', 'OPEN', 'opened_at', now)
        return {0, 'OPEN'}
    end
end

return {1, 'CLOSED'}
`

// Lua script to record result
var recordResultScript = `
local state_key = KEYS[1]
local window_key = KEYS[2]
local success = tonumber(ARGV[1])
local window_size = tonumber(ARGV[2])
local half_open_successes = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

-- Get state
local state = redis.call('HGET', state_key, 'state') or 'CLOSED'

-- Update window
redis.call('HINCRBY', window_key, 'total', 1)
if success == 0 then
    redis.call('HINCRBY', window_key, 'failures', 1)
end
redis.call('EXPIRE', window_key, window_size * 2)

-- Handle HALF_OPEN state
if state == 'HALF_OPEN' then
    if success == 0 then
        -- Failure in HALF_OPEN: reopen circuit
        redis.call('HSET', state_key, 'state', 'OPEN', 'opened_at', now)
    else
        -- Success in HALF_OPEN: count it
        local successes = redis.call('HINCRBY', state_key, 'successes', 1)
        if successes >= half_open_successes then
            redis.call('DEL', state_key)  -- Back to CLOSED
        end
    end
end

return 1
`

// Breaker implements circuit breaker pattern with Redis
type Breaker struct {
	redis   redis.Evaluator
	service string
}

// NewBreaker creates a circuit breaker for a service
func NewBreaker(r redis.Evaluator, service string) *Breaker {
	return &Breaker{redis: r, service: service}
}

// Result of circuit breaker check
type Result struct {
	Allowed bool
	State   State
}

// Allow checks if request should be allowed
func (b *Breaker) Allow(ctx context.Context) Result {
	now := time.Now().Unix()
	windowStart := (now / WindowSize) * WindowSize

	stateKey := "circuit:" + b.service + ":state"
	windowKey := "circuit:" + b.service + ":window:" + itoa(windowStart)

	result, err := b.redis.Eval(ctx, circuitBreakerScript,
		[]string{stateKey, windowKey},
		now, CooldownSeconds, WindowSize, FailureThreshold, FailureRatePercent, HalfOpenSuccesses,
	)

	// Fail-open if Redis unavailable
	if err != nil {
		return Result{Allowed: true, State: StateClosed}
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) < 2 {
		return Result{Allowed: true, State: StateClosed}
	}

	allowed := toInt(arr[0]) == 1
	state := State(toString(arr[1]))

	return Result{Allowed: allowed, State: state}
}

// RecordSuccess records a successful request
func (b *Breaker) RecordSuccess(ctx context.Context) {
	b.recordResult(ctx, true)
}

// RecordFailure records a failed request
func (b *Breaker) RecordFailure(ctx context.Context) {
	b.recordResult(ctx, false)
}

func (b *Breaker) recordResult(ctx context.Context, success bool) {
	now := time.Now().Unix()
	windowStart := (now / WindowSize) * WindowSize

	stateKey := "circuit:" + b.service + ":state"
	windowKey := "circuit:" + b.service + ":window:" + itoa(windowStart)

	successInt := 0
	if success {
		successInt = 1
	}

	b.redis.Eval(ctx, recordResultScript,
		[]string{stateKey, windowKey},
		successInt, WindowSize, HalfOpenSuccesses, now,
	)
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func toInt(v interface{}) int {
	if n, ok := v.(int64); ok {
		return int(n)
	}
	return 0
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
