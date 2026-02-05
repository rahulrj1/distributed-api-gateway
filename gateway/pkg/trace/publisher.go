package trace

import (
	"context"
	"fmt"
	"log"

	goredis "github.com/redis/go-redis/v9"
)

// Publisher handles publishing trace events to Redis Pub/Sub.
type Publisher struct {
	client  *goredis.Client
	enabled bool
}

// NewPublisher creates a new trace publisher.
// If client is nil, publishing is disabled (events are silently dropped).
func NewPublisher(client *goredis.Client) *Publisher {
	return &Publisher{
		client:  client,
		enabled: client != nil,
	}
}

// Publish sends a trace event to Redis Pub/Sub.
// Channel format: trace:{traceId}
func (p *Publisher) Publish(ctx context.Context, event *Event) error {
	if !p.enabled {
		return nil
	}

	data, err := event.JSON()
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	channel := fmt.Sprintf("trace:%s", event.TraceID)
	if err := p.client.Publish(ctx, channel, data).Err(); err != nil {
		log.Printf("trace publish error: %v", err)
		return err
	}

	return nil
}

// PublishRequest publishes the initial request info.
func (p *Publisher) PublishRequest(ctx context.Context, info *RequestInfo) error {
	if !p.enabled {
		return nil
	}

	// Create initial event with request details
	event := NewEvent(info.TraceID, StepReceived, StatusSuccess).
		WithDetail("method", info.Method).
		WithDetail("path", info.Path).
		WithDetail("service", info.Service)

	if info.ClientID != "" {
		event.WithDetail("client_id", info.ClientID)
	}

	return p.Publish(ctx, event)
}

// Enabled returns whether tracing is enabled.
func (p *Publisher) Enabled() bool {
	return p.enabled
}
