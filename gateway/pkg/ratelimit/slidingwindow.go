package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/distributed-api-gateway/gateway/pkg/redis"
)

const (
	WindowSize   = 60 // seconds
	DefaultLimit = 100
)

// Lua script for atomic sliding window rate limiting
// Keys: [current_window_key, prev_window_key]
// Args: [limit, current_window_start, now, window_size]
// Returns: [allowed (0/1), current_count, retry_after]
var slidingWindowScript = `
local curr_key = KEYS[1]
local prev_key = KEYS[2]
local limit = tonumber(ARGV[1])
local window_start = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window_size = tonumber(ARGV[4])

-- Get counts
local curr_count = tonumber(redis.call('GET', curr_key) or '0')
local prev_count = tonumber(redis.call('GET', prev_key) or '0')

-- Calculate weighted count using sliding window
local elapsed = now - window_start
local weight = 1 - (elapsed / window_size)
if weight < 0 then weight = 0 end
local weighted_count = (prev_count * weight) + curr_count

-- Check limit
if weighted_count >= limit then
    local retry_after = window_size - elapsed
    if retry_after < 1 then retry_after = 1 end
    return {0, weighted_count, retry_after}
end

-- Increment and set TTL
redis.call('INCR', curr_key)
redis.call('EXPIRE', curr_key, window_size * 2)

return {1, weighted_count + 1, 0}
`

// Limiter implements sliding window rate limiting with Redis
type Limiter struct {
	redis redis.Evaluator
	limit int
}

// NewLimiter creates a rate limiter
func NewLimiter(r redis.Evaluator, limit int) *Limiter {
	if limit <= 0 {
		limit = DefaultLimit
	}
	return &Limiter{redis: r, limit: limit}
}

// Result of a rate limit check
type Result struct {
	Allowed    bool
	Count      int
	RetryAfter int // seconds until reset
}

// Allow checks if request is allowed. Key is typically client_id or IP.
func (l *Limiter) Allow(ctx context.Context, key string) Result {
	now := time.Now().Unix()
	windowStart := (now / WindowSize) * WindowSize
	prevWindowStart := windowStart - WindowSize

	currKey := "ratelimit:" + key + ":" + strconv.FormatInt(windowStart, 10)
	prevKey := "ratelimit:" + key + ":" + strconv.FormatInt(prevWindowStart, 10)

	result, err := l.redis.Eval(ctx, slidingWindowScript,
		[]string{currKey, prevKey},
		l.limit, windowStart, now, WindowSize,
	)

	// Fail-open: if Redis fails, allow request
	if err != nil {
		return Result{Allowed: true}
	}

	// Parse result: [allowed, count, retry_after]
	arr, ok := result.([]interface{})
	if !ok || len(arr) < 3 {
		return Result{Allowed: true}
	}

	allowed := toInt(arr[0]) == 1
	count := toInt(arr[1])
	retryAfter := toInt(arr[2])

	return Result{Allowed: allowed, Count: count, RetryAfter: retryAfter}
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
