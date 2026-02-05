package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	goredis "github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev; restrict in production
	},
}

// TraceWebSocket handles WebSocket connections for trace streaming.
// Endpoint: GET /ws/trace/{traceId}
func TraceWebSocket(redisClient *goredis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract trace ID from path: /ws/trace/{traceId}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			http.Error(w, "trace ID required", http.StatusBadRequest)
			return
		}
		traceID := parts[2]

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Subscribe to trace channel
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		channel := "trace:" + traceID
		pubsub := redisClient.Subscribe(ctx, channel)
		defer pubsub.Close()

		// Wait for subscription confirmation
		_, err = pubsub.Receive(ctx)
		if err != nil {
			log.Printf("Redis subscribe error: %v", err)
			conn.WriteJSON(map[string]string{"error": "subscription failed"})
			return
		}

		// Send confirmation
		conn.WriteJSON(map[string]interface{}{
			"type":     "subscribed",
			"trace_id": traceID,
			"channel":  channel,
		})

		// Handle incoming messages from Redis
		ch := pubsub.Channel()
		
		// Set up ping/pong for keepalive
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Start goroutine to handle client messages (pings, close)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()

		// Ping ticker
		pingTicker := time.NewTicker(30 * time.Second)
		defer pingTicker.Stop()

		// Timeout after 5 minutes of no messages
		timeout := time.NewTimer(5 * time.Minute)
		defer timeout.Stop()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// Forward trace event to WebSocket
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					log.Printf("WebSocket write error: %v", err)
					return
				}
				// Reset timeout on activity
				timeout.Reset(5 * time.Minute)

			case <-pingTicker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}

			case <-timeout.C:
				conn.WriteJSON(map[string]string{"type": "timeout", "message": "connection timed out"})
				return

			case <-done:
				return

			case <-ctx.Done():
				return
			}
		}
	}
}
