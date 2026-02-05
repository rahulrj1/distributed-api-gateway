package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Evaluator runs Lua scripts (interface for mocking in tests)
type Evaluator interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
}

// Client wraps Redis operations needed by the gateway
type Client struct {
	rdb *redis.Client
}

var _ Evaluator = (*Client)(nil)

// New creates a Redis client
func New(addr string) *Client {
	return &Client{
		rdb: redis.NewClient(&redis.Options{
			Addr:         addr,
			DialTimeout:  1 * time.Second,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		}),
	}
}

// Ping checks Redis connectivity
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Eval runs a Lua script
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.rdb.Eval(ctx, script, keys, args...).Result()
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}
